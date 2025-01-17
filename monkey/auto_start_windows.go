// +build windows

package monkey

import (
	"dxkite.cn/log"
	"golang.org/x/sys/windows/registry"
)

func AutoStart(cmd string) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.ALL_ACCESS)
	if err != nil {
		log.Println(err)
		return
	}
	raw, _, _ := k.GetStringValue("Mino")
	if raw == cmd {
		log.Println("auto start is set", cmd)
		return
	}
	if err := k.SetStringValue("Mino", cmd); err != nil {
		log.Warn("set auto start error", err)
		return
	}
	log.Println("auto start is set", cmd)
}
