package memory

//#include <memory.h>
import "C"

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"unsafe"

	"golang.org/x/sys/windows"
)

func Main() {
	if len(os.Args) < 3 {
		fmt.Println(OutputErrorPrefix + "Invalid arguments")
		fmt.Printf("Usage: %s <size> <native_memprotect> [file]\n", os.Args[0])
		os.Exit(1)
	}

	filename := ""
	if len(os.Args) >= 4 {
		filename = os.Args[3]
	}

	size, err := strconv.ParseUint(os.Args[1], 10, 64)
	if err != nil {
		fmt.Printf(OutputErrorPrefix+"Invalid size value, %v\n", err)
		os.Exit(1)
	}

	protect, err := strconv.ParseUint(os.Args[2], 10, 32)
	if err != nil {
		fmt.Printf(OutputErrorPrefix+"Invalid protect value, %v\n", err)
		os.Exit(1)
	}

	addr, err := windows.VirtualAlloc(0, uintptr(size), windows.MEM_RESERVE|windows.MEM_COMMIT, windows.PAGE_READWRITE)
	if err != nil {
		fmt.Printf(OutputErrorPrefix+"Could not alloc, reason: %v\n", err)
		os.Exit(5)
	}
	defer func() {
		windows.VirtualFree(addr, 0, windows.MEM_RELEASE)
	}()

	var data []byte

	if filename != "" {
		f, err := os.Open(filename)
		if err != nil {
			fmt.Printf(OutputErrorPrefix+"Could not open file, reason: %v\n", err)
			os.Exit(2)
		}
		data, err = ioutil.ReadAll(f)
		if err != nil {
			fmt.Printf(OutputErrorPrefix+"Could not read from file, reason: %v\n", err)
			os.Exit(3)
		}
		f.Close()

		size = uint64(len(data))
	} else {
		fmt.Println(OutputReady)
		data, err = ioutil.ReadAll(io.LimitReader(os.Stdin, int64(size)))
		if err != nil {
			fmt.Printf(OutputErrorPrefix+"Could not read from stdin, reason: %v\n", err)
			os.Exit(4)
		}
	}

	C.memcpy(unsafe.Pointer(addr), unsafe.Pointer(&data[0]), C.size_t(size))

	var oldProtect uint32
	err = windows.VirtualProtect(addr, uintptr(len(data)), uint32(protect), &oldProtect)
	if err != nil {
		fmt.Printf(OutputErrorPrefix+"Failed to set protect, reason: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf(OutputAddressPrefix+"%d\n", addr)

	if filename != "" {
		fmt.Println("Press Enter to close application...")
		// Wait for user enter
		fmt.Scanln()
	} else {
		// Wait for stdin close
		ioutil.ReadAll(os.Stdin)
	}
}
