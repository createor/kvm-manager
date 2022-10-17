package utils

import (
	"fmt"
)

// 创建qcow2磁盘
func NewDisk(size, path string) string {
	return fmt.Sprintf("qemu-img create -f qcow2 -o size=%s %s", size, path)
}
