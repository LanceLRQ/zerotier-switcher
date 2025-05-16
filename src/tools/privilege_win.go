//go:build windows
// +build windows

package tools

import (
	"fmt"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
	"os"
	"time"
)

func IsRunAsRoot() bool {
	var sid *windows.SID

	// 虽然这个函数会返回错误，但我们忽略它
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		fmt.Printf("SID Error: %s\n", err)
		return false
	}
	defer windows.FreeSid(sid)

	// 这个token是0，表示我们检查当前线程/进程的token
	token := windows.Token(0)

	member, err := token.IsMember(sid)
	if err != nil {
		fmt.Printf("Token Membership Error: %s\n", err)
		return false
	}

	return member
}
