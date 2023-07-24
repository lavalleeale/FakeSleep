package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"unsafe"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Needs time and file name")
		os.Exit(1)
	}
	time, err := strconv.ParseUint(os.Args[1], 10, 64)
	if err != nil {
		log.Panicf("Invalid time: %s\n", os.Args[1])
	}
	test := exec.Command(os.Args[2])
	test.Stdout = os.Stdout
	test.Stderr = os.Stderr
	test.SysProcAttr = &syscall.SysProcAttr{Ptrace: true}
	test.Start()
	test.Wait()
	pid := test.Process.Pid
	var wopt syscall.WaitStatus
	regs := &syscall.PtraceRegs{}
	for {
		err := syscall.PtraceSyscall(pid, 0)
		if err != nil {
			os.Exit(0)
		}
		syscall.Wait4(pid, &wopt, 0, nil)
		syscall.PtraceGetRegs(pid, regs)
		if err != nil {
			os.Exit(0)
		}
		switch regs.Orig_rax {
		case syscall.SYS_TIME:
			syscall.PtraceSyscall(pid, 0)
			syscall.Wait4(pid, &wopt, 0, nil)
			syscall.PtraceGetRegs(pid, regs)
			fmt.Printf("Original time response: %d\n", regs.Rax)
			regs.Rax = time
			err = syscall.PtraceSetRegs(pid, regs)
			if err != nil {
				panic(err)
			}
		case syscall.SYS_CLOCK_NANOSLEEP:
			var timespec syscall.Timespec
			var data = make([]byte, unsafe.Sizeof(timespec))
			_, err = syscall.PtracePeekData(pid, uintptr(regs.Rdx), data)
			if err != nil {
				panic(err)
			}
			if getEndian() {
				binary.Read(bytes.NewBuffer(data), binary.BigEndian, &timespec)
			} else {
				binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &timespec)
			}
			fmt.Printf("Original sleep duration: %d\n", timespec.Nano()/1_000_000_000)
			time += uint64(timespec.Nano() / 1_000_000_000)
			_, err = syscall.PtracePokeData(pid, uintptr(regs.Rdx), make([]byte, 16))
			syscall.PtraceSyscall(pid, 0)
			syscall.Wait4(pid, &wopt, 0, nil)
			if err != nil {
				panic(err)
			}
		default:
		}
	}
}

const INT_SIZE int = int(unsafe.Sizeof(0))

// true = big endian, false = little endian
func getEndian() (ret bool) {
	var i int = 0x1
	bs := (*[INT_SIZE]byte)(unsafe.Pointer(&i))
	if bs[0] == 0 {
		return true
	} else {
		return false
	}

}
