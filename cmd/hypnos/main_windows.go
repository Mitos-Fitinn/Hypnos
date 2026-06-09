package main

import (
	_ "embed"
	"encoding/binary"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

//go:embed assets/hypnos-active.ico
var activeIconData []byte

//go:embed assets/hypnos-inactive.ico
var inactiveIconData []byte

const (
	appName = "Hypnos"

	wmCreate         = 0x0001
	wmDestroy        = 0x0002
	wmClose          = 0x0010
	wmCommand        = 0x0111
	wmPaint          = 0x000F
	wmSetFont        = 0x0030
	wmCtlColorEdit   = 0x0133
	wmCtlColorBtn    = 0x0135
	wmCtlColorStatic = 0x0138
	wmKeyDown        = 0x0100
	wmChar           = 0x0102
	wmTimer          = 0x0113
	wmApp            = 0x8000
	wmPowerUpdate    = wmApp + 2
	wmLButtonDown    = 0x0201
	wmRButtonUp      = 0x0205
	wmContextMenu    = 0x007B

	nimAdd    = 0x00000000
	nimModify = 0x00000001
	nimDelete = 0x00000002

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004

	idiApplication = 32512
	idcArrow       = 32512
	colorWindow    = 5
	logPixelsX     = 88

	mfString    = 0x00000000
	mfPopup     = 0x00000010
	mfSeparator = 0x00000800

	tpmRightButton = 0x0002
	tpmReturnCmd   = 0x0100

	swShow   = 5
	vkReturn = 0x0D
	vkEscape = 0x1B

	wsCaption = 0x00C00000
	wsSysMenu = 0x00080000
	wsVisible = 0x10000000
	wsChild   = 0x40000000
	wsTabStop = 0x00010000
	wsBorder  = 0x00800000
	wsVScroll = 0x00200000

	dtsUpDown        = 0x0001
	dtsTimeFormat    = 0x0009
	dtmFirst         = 0x1000
	dtmGetSystemTime = dtmFirst + 1
	dtmSetSystemTime = dtmFirst + 2
	dtmSetFormatW    = dtmFirst + 50
	gdtValid         = 0
	iccDateClasses   = 0x00000100

	bsDefPushButton = 0x00000001
	bsAutoCheckbox  = 0x00000003

	esLeft           = 0x00000000
	esCenter         = 0x00000001
	esMultiline      = 0x00000004
	esReadOnly       = 0x00000800
	esAutoHScroll    = 0x00000080
	esAutoVScroll    = 0x00000040
	bmGetCheck       = 0x00F0
	bmSetCheck       = 0x00F1
	bstChecked       = 1
	mbIconError      = 0x00000010
	defaultCharset   = 1
	cleartypeQuality = 5
	transparentBk    = 1

	dtLeft       = 0x00000000
	dtCenter     = 0x00000001
	dtVCenter    = 0x00000004
	dtWordBreak  = 0x00000010
	dtSingleLine = 0x00000020

	timerID    = 1
	intervalMs = 59000

	inputKeyboard  = 1
	keyEventFKeyUp = 0x0002
	vkF24          = 0x87

	esSystemRequired = 0x00000001
	esContinuous     = 0x80000000

	turnOnCommand   = 1001
	turnOffCommand  = 1002
	settingsCommand = 1003
	exitCommand     = 1004
	aboutCommand    = 1005

	quickTimerCommandBase = 1100

	timeEditID      = 2001
	shutdownCheckID = 2002
	startButtonID   = 2003
	cancelButtonID  = 2004
	aboutOkButtonID = 2005
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")

	procRegisterClassExW         = user32.NewProc("RegisterClassExW")
	procCreateWindowExW          = user32.NewProc("CreateWindowExW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procDestroyWindow            = user32.NewProc("DestroyWindow")
	procIsWindow                 = user32.NewProc("IsWindow")
	procPostQuitMessage          = user32.NewProc("PostQuitMessage")
	procPostMessageW             = user32.NewProc("PostMessageW")
	procGetMessageW              = user32.NewProc("GetMessageW")
	procTranslateMessage         = user32.NewProc("TranslateMessage")
	procDispatchMessageW         = user32.NewProc("DispatchMessageW")
	procBeginPaint               = user32.NewProc("BeginPaint")
	procEndPaint                 = user32.NewProc("EndPaint")
	procInvalidateRect           = user32.NewProc("InvalidateRect")
	procMessageBoxW              = user32.NewProc("MessageBoxW")
	procLoadIconW                = user32.NewProc("LoadIconW")
	procLoadCursorW              = user32.NewProc("LoadCursorW")
	procSetFocus                 = user32.NewProc("SetFocus")
	procSetProcessDPIAware       = user32.NewProc("SetProcessDPIAware")
	procGetDC                    = user32.NewProc("GetDC")
	procReleaseDC                = user32.NewProc("ReleaseDC")
	procSetTimer                 = user32.NewProc("SetTimer")
	procKillTimer                = user32.NewProc("KillTimer")
	procCreatePopupMenu          = user32.NewProc("CreatePopupMenu")
	procAppendMenuW              = user32.NewProc("AppendMenuW")
	procTrackPopupMenu           = user32.NewProc("TrackPopupMenu")
	procDestroyMenu              = user32.NewProc("DestroyMenu")
	procGetCursorPos             = user32.NewProc("GetCursorPos")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procSendInput                = user32.NewProc("SendInput")
	procSendMessageW             = user32.NewProc("SendMessageW")
	procGetDlgItem               = user32.NewProc("GetDlgItem")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procShowWindow               = user32.NewProc("ShowWindow")
	procUpdateWindow             = user32.NewProc("UpdateWindow")
	procSetWindowTextW           = user32.NewProc("SetWindowTextW")
	procCreateIconFromResourceEx = user32.NewProc("CreateIconFromResourceEx")
	procShellNotifyIconW         = shell32.NewProc("Shell_NotifyIconW")
	procGetModuleHandleW         = kernel32.NewProc("GetModuleHandleW")
	procSetThreadExecutionState  = kernel32.NewProc("SetThreadExecutionState")
	procCreateFontW              = gdi32.NewProc("CreateFontW")
	procCreateSolidBrush         = gdi32.NewProc("CreateSolidBrush")
	procCreatePen                = gdi32.NewProc("CreatePen")
	procDeleteObject             = gdi32.NewProc("DeleteObject")
	procSelectObject             = gdi32.NewProc("SelectObject")
	procSetBkMode                = gdi32.NewProc("SetBkMode")
	procSetBkColor               = gdi32.NewProc("SetBkColor")
	procSetTextColor             = gdi32.NewProc("SetTextColor")
	procDrawTextW                = user32.NewProc("DrawTextW")
	procRoundRect                = gdi32.NewProc("RoundRect")
	procGetDeviceCaps            = gdi32.NewProc("GetDeviceCaps")
	procInitCommonControlsEx     = comctl32.NewProc("InitCommonControlsEx")

	hInstance       uintptr
	mainWindow      uintptr
	settingsWindow  uintptr
	aboutWindow     uintptr
	mainWndProc     uintptr
	settingsWndProc uintptr
	aboutWndProc    uintptr
	activeIcon      uintptr
	inactiveIcon    uintptr
	uiFont          uintptr
	titleFont       uintptr
	timeFont        uintptr
	editBrush       uintptr
	windowBrush     uintptr
	dpiScale        int
	trayMessage     = uint32(wmApp + 1)

	active                = true
	activeUntil           time.Time
	shutdownWhenOff       bool
	settingsTimeText      string
	settingsDraftShutdown bool

	quickTimers = []struct {
		label   string
		minutes int
	}{
		{"30 min", 30},
		{"1 h", 60},
		{"1 h 30 min", 90},
		{"2 h", 120},
		{"2 h 30 min", 150},
		{"3 h", 180},
	}
)

type point struct {
	x int32
	y int32
}

type rect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

type paintStruct struct {
	hdc         uintptr
	erase       int32
	rcPaint     rect
	restore     int32
	incUpdate   int32
	rgbReserved [32]byte
}

type msg struct {
	hWnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type notifyIconData struct {
	cbSize           uint32
	hWnd             uintptr
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	hIcon            uintptr
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uVersion         uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         [16]byte
	hBalloonIcon     uintptr
}

type keyboardInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

type input struct {
	inputType uint32
	padding   uint32
	ki        keyboardInput
	unionPad  [8]byte
}

type initCommonControlsEx struct {
	dwSize uint32
	dwICC  uint32
}

type systemTime struct {
	year         uint16
	month        uint16
	dayOfWeek    uint16
	day          uint16
	hour         uint16
	minute       uint16
	second       uint16
	milliseconds uint16
}

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	procSetProcessDPIAware.Call()
	initCommonControls()

	hInstance, _, _ = procGetModuleHandleW.Call(0)
	dpiScale = currentDPIScale()
	activeIcon = iconFromICO(activeIconData)
	inactiveIcon = iconFromICO(inactiveIconData)
	if activeIcon == 0 {
		activeIcon, _, _ = procLoadIconW.Call(0, idiApplication)
	}
	if inactiveIcon == 0 {
		inactiveIcon = activeIcon
	}
	uiFont = createFont(scale(18), 400, "Segoe UI")
	titleFont = createFont(scale(24), 600, "Segoe UI")
	timeFont = createFont(scale(22), 500, "Segoe UI")
	editBrush, _, _ = procCreateSolidBrush.Call(color(24, 28, 38))
	windowBrush, _, _ = procCreateSolidBrush.Call(color(255, 255, 255))

	mainWndProc = syscall.NewCallback(mainWindowProc)
	settingsWndProc = syscall.NewCallback(settingsWindowProc)
	aboutWndProc = syscall.NewCallback(aboutWindowProc)

	registerWindowClass(appName+"MainWindow", mainWndProc)
	registerWindowClass(appName+"SettingsWindow", settingsWndProc)
	registerWindowClass(appName+"AboutWindow", aboutWndProc)

	mainWindow, _, _ = procCreateWindowExW.Call(
		0,
		ptr(appName+"MainWindow"),
		ptr(appName),
		0,
		0, 0, 0, 0,
		0, 0, hInstance, 0,
	)
	if mainWindow == 0 {
		return
	}

	addTrayIcon()
	syncPowerState()
	sendF24()
	procSetTimer.Call(mainWindow, timerID, intervalMs, 0)

	var message msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&message)))
	}
}

func registerWindowClass(name string, wndProc uintptr) {
	cursor, _, _ := procLoadCursorW.Call(0, idcArrow)
	class := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   wndProc,
		hInstance:     hInstance,
		hIcon:         activeIcon,
		hCursor:       cursor,
		hbrBackground: colorWindow + 1,
		lpszClassName: utf16Ptr(name),
		hIconSm:       activeIcon,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&class)))
}

func mainWindowProc(hWnd uintptr, message uint32, wParam uintptr, lParam uintptr) uintptr {
	switch message {
	case wmTimer:
		tick()
		return 0
	case wmPowerUpdate:
		syncPowerState()
		return 0
	case wmCommand:
		handleCommand(wParam)
		return 0
	case trayMessage:
		if uint32(lParam) == wmRButtonUp || uint32(lParam) == wmContextMenu {
			showMenu()
		}
		return 0
	case wmDestroy:
		procKillTimer.Call(hWnd, timerID)
		allowSystemSleep()
		removeTrayIcon()
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hWnd, uintptr(message), wParam, lParam)
	return ret
}

func settingsWindowProc(hWnd uintptr, message uint32, wParam uintptr, lParam uintptr) uintptr {
	switch message {
	case wmCreate:
		createSettingsControls(hWnd)
		return 0
	case wmCtlColorEdit, wmCtlColorStatic, wmCtlColorBtn:
		procSetTextColor.Call(wParam, color(0, 0, 0))
		procSetBkColor.Call(wParam, color(255, 255, 255))
		return windowBrush
	case wmCommand:
		switch lowWord(wParam) {
		case startButtonID:
			if applySettings(hWnd) {
				procDestroyWindow.Call(hWnd)
			}
			return 0
		case cancelButtonID:
			procDestroyWindow.Call(hWnd)
			return 0
		}
	case wmClose:
		procDestroyWindow.Call(hWnd)
		return 0
	case wmDestroy:
		if settingsWindow == hWnd {
			settingsWindow = 0
		}
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hWnd, uintptr(message), wParam, lParam)
	return ret
}

func aboutWindowProc(hWnd uintptr, message uint32, wParam uintptr, lParam uintptr) uintptr {
	switch message {
	case wmCreate:
		createAboutControls(hWnd)
		return 0
	case wmCtlColorEdit, wmCtlColorStatic, wmCtlColorBtn:
		procSetTextColor.Call(wParam, color(0, 0, 0))
		procSetBkColor.Call(wParam, color(255, 255, 255))
		return windowBrush
	case wmCommand:
		if lowWord(wParam) == aboutOkButtonID {
			procDestroyWindow.Call(hWnd)
			return 0
		}
	case wmClose:
		procDestroyWindow.Call(hWnd)
		return 0
	case wmDestroy:
		if aboutWindow == hWnd {
			aboutWindow = 0
		}
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hWnd, uintptr(message), wParam, lParam)
	return ret
}

func createSettingsControls(hWnd uintptr) {
	defaultUntil := time.Now().Add(30 * time.Minute)
	if !activeUntil.IsZero() {
		defaultUntil = activeUntil
	}

	setFont(createControl("STATIC", "Settings", wsChild|wsVisible, 28, 24, 380, 34, hWnd, 0), titleFont)
	setFont(createControl("STATIC", "Keep Hypnos active until this time.", wsChild|wsVisible, 28, 64, 390, 26, hWnd, 0), uiFont)
	timePicker := createControl("SysDateTimePick32", "", wsChild|wsVisible|wsTabStop|dtsTimeFormat|dtsUpDown, 28, 106, 126, 32, hWnd, timeEditID)
	setFont(timePicker, uiFont)
	setTimePicker(timePicker, defaultUntil)
	setFont(createControl("STATIC", "Format: 23:30", wsChild|wsVisible, 176, 110, 170, 26, hWnd, 0), uiFont)

	check := createControl("BUTTON", "Shut down PC when Hypnos turns off", wsChild|wsVisible|wsTabStop|bsAutoCheckbox, 28, 166, 360, 32, hWnd, shutdownCheckID)
	setFont(check, uiFont)
	if shutdownWhenOff {
		procSendMessage(check, bmSetCheck, bstChecked, 0)
	}

	setFont(createControl("BUTTON", "Start", wsChild|wsVisible|wsTabStop|bsDefPushButton, 190, 226, 96, 38, hWnd, startButtonID), uiFont)
	setFont(createControl("BUTTON", "Cancel", wsChild|wsVisible|wsTabStop, 300, 226, 116, 38, hWnd, cancelButtonID), uiFont)
}

func createAboutControls(hWnd uintptr) {
	setFont(createControl("STATIC", appName, wsChild|wsVisible, 30, 26, 460, 34, hWnd, 0), titleFont)
	setFont(createControl("STATIC", "About it", wsChild|wsVisible, 30, 70, 460, 26, hWnd, 0), uiFont)
	setFont(createControl("EDIT", strings.ReplaceAll(aboutProjectText, "\n", "\r\n"), wsChild|wsVisible|wsBorder|esMultiline|esReadOnly|esAutoVScroll|wsVScroll, 30, 112, 500, 190, hWnd, 0), uiFont)
	setFont(createControl("STATIC", aboutWatermark, wsChild|wsVisible, 30, 326, 360, 26, hWnd, 0), uiFont)
	setFont(createControl("BUTTON", "OK", wsChild|wsVisible|wsTabStop|bsDefPushButton, 430, 322, 100, 38, hWnd, aboutOkButtonID), uiFont)
}

func createControl(className string, text string, style uintptr, x int, y int, width int, height int, parent uintptr, id uintptr) uintptr {
	hWnd, _, _ := procCreateWindowExW.Call(
		0,
		ptr(className),
		ptr(text),
		style,
		uintptr(scale(x)),
		uintptr(scale(y)),
		uintptr(scale(width)),
		uintptr(scale(height)),
		parent,
		id,
		hInstance,
		0,
	)
	return hWnd
}

func initCommonControls() {
	controls := initCommonControlsEx{
		dwSize: uint32(unsafe.Sizeof(initCommonControlsEx{})),
		dwICC:  iccDateClasses,
	}
	procInitCommonControlsEx.Call(uintptr(unsafe.Pointer(&controls)))
}

func setTimePicker(hWnd uintptr, value time.Time) {
	procSendMessage(hWnd, dtmSetFormatW, 0, ptr("HH':'mm"))
	selected := systemTime{
		year:   uint16(value.Year()),
		month:  uint16(value.Month()),
		day:    uint16(value.Day()),
		hour:   uint16(value.Hour()),
		minute: uint16(value.Minute()),
	}
	procSendMessage(hWnd, dtmSetSystemTime, gdtValid, uintptr(unsafe.Pointer(&selected)))
}

func createFont(height int, weight int, face string) uintptr {
	font, _, _ := procCreateFontW.Call(
		uintptr(height),
		0,
		0,
		0,
		uintptr(weight),
		0,
		0,
		0,
		defaultCharset,
		0,
		0,
		cleartypeQuality,
		0,
		ptr(face),
	)
	return font
}

func setFont(hWnd uintptr, font uintptr) {
	if hWnd == 0 || font == 0 {
		return
	}
	procSendMessage(hWnd, wmSetFont, font, 1)
}

func drawButton(hdc uintptr, bounds rect, text string, fill uintptr) {
	drawRound(hdc, bounds, fill, fill, 12)
	drawText(hdc, text, bounds, uiFont, color(255, 255, 255), dtCenter|dtVCenter|dtSingleLine)
}

func drawRound(hdc uintptr, bounds rect, fill uintptr, border uintptr, radius int32) {
	brush, _, _ := procCreateSolidBrush.Call(fill)
	pen, _, _ := procCreatePen.Call(0, 1, border)
	defer procDeleteObject.Call(brush)
	defer procDeleteObject.Call(pen)

	oldBrush, _, _ := procSelectObject.Call(hdc, brush)
	oldPen, _, _ := procSelectObject.Call(hdc, pen)
	defer procSelectObject.Call(hdc, oldBrush)
	defer procSelectObject.Call(hdc, oldPen)

	if radius <= 0 {
		radius = 1
	}
	procRoundRect.Call(
		hdc,
		uintptr(bounds.left),
		uintptr(bounds.top),
		uintptr(bounds.right),
		uintptr(bounds.bottom),
		uintptr(radius),
		uintptr(radius),
	)
}

func drawText(hdc uintptr, text string, bounds rect, font uintptr, textColor uintptr, flags uintptr) {
	if font != 0 {
		oldFont, _, _ := procSelectObject.Call(hdc, font)
		defer procSelectObject.Call(hdc, oldFont)
	}
	procSetBkMode.Call(hdc, transparentBk)
	procSetTextColor.Call(hdc, textColor)
	textPtr := utf16Ptr(text)
	procDrawTextW.Call(
		hdc,
		uintptr(unsafe.Pointer(textPtr)),
		^uintptr(0),
		uintptr(unsafe.Pointer(&bounds)),
		flags,
	)
}

func pointFromLParam(lParam uintptr) point {
	return point{
		x: int32(int16(lParam & 0xffff)),
		y: int32(int16((lParam >> 16) & 0xffff)),
	}
}

func inRect(pt point, bounds rect) bool {
	return pt.x >= bounds.left && pt.x <= bounds.right && pt.y >= bounds.top && pt.y <= bounds.bottom
}

func color(red int, green int, blue int) uintptr {
	return uintptr(red | green<<8 | blue<<16)
}

func currentDPIScale() int {
	hdc, _, _ := procGetDC.Call(0)
	if hdc == 0 {
		return 100
	}
	defer procReleaseDC.Call(0, hdc)

	dpi, _, _ := procGetDeviceCaps.Call(hdc, logPixelsX)
	if dpi == 0 {
		return 100
	}
	return int((dpi*100 + 48) / 96)
}

func scale(value int) int {
	if dpiScale <= 0 {
		return value
	}
	return (value*dpiScale + 50) / 100
}

func procSendMessage(hWnd uintptr, message uint32, wParam uintptr, lParam uintptr) uintptr {
	ret, _, _ := procSendMessageW.Call(hWnd, uintptr(message), wParam, lParam)
	return ret
}

func applySettings(hWnd uintptr) bool {
	timePicker, _, _ := procGetDlgItem.Call(hWnd, timeEditID)
	check, _, _ := procGetDlgItem.Call(hWnd, shutdownCheckID)
	until, ok := timePickerValue(timePicker)
	if !ok {
		procMessageBoxW.Call(
			hWnd,
			ptr("Select a valid time."),
			ptr(appName),
			mbIconError,
		)
		return false
	}

	active = true
	activeUntil = until
	shutdownWhenOff = procSendMessage(check, bmGetCheck, 0, 0) == bstChecked
	notifyPowerStateChanged()
	updateTrayIcon()
	sendF24()
	return true
}

func timePickerValue(hWnd uintptr) (time.Time, bool) {
	var selected systemTime
	ret := procSendMessage(hWnd, dtmGetSystemTime, 0, uintptr(unsafe.Pointer(&selected)))
	if ret != gdtValid {
		return time.Time{}, false
	}

	now := time.Now()
	until := time.Date(now.Year(), now.Month(), now.Day(), int(selected.hour), int(selected.minute), 0, 0, time.Local)
	if !until.After(now) {
		until = until.Add(24 * time.Hour)
	}
	return until, true
}

func paintSettings(hWnd uintptr) {
	var ps paintStruct
	hdc, _, _ := procBeginPaint.Call(hWnd, uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(hWnd, uintptr(unsafe.Pointer(&ps)))

	drawRound(hdc, rect{left: 0, top: 0, right: 470, bottom: 330}, color(15, 17, 23), color(15, 17, 23), 0)
	drawText(hdc, "Settings", rect{left: 28, top: 22, right: 430, bottom: 56}, titleFont, color(239, 244, 255), dtLeft|dtSingleLine)
	drawText(hdc, "Keep Hypnos active until this time.", rect{left: 28, top: 58, right: 430, bottom: 84}, uiFont, color(150, 160, 178), dtLeft|dtSingleLine)

	drawRound(hdc, rect{left: 28, top: 102, right: 190, bottom: 166}, color(24, 28, 38), color(74, 144, 226), 14)
	drawText(hdc, "Format: 23:30", rect{left: 212, top: 122, right: 390, bottom: 150}, uiFont, color(150, 160, 178), dtLeft|dtSingleLine)

	drawRound(hdc, rect{left: 28, top: 190, right: 52, bottom: 214}, color(24, 28, 38), color(74, 144, 226), 6)
	if settingsDraftShutdown {
		drawRound(hdc, rect{left: 34, top: 196, right: 46, bottom: 208}, color(74, 144, 226), color(74, 144, 226), 4)
	}
	drawText(hdc, "Shut down PC when Hypnos turns off", rect{left: 66, top: 190, right: 430, bottom: 218}, uiFont, color(226, 232, 240), dtLeft|dtSingleLine)

	drawButton(hdc, rect{left: 208, top: 252, right: 306, bottom: 292}, "Start", color(37, 99, 235))
	drawButton(hdc, rect{left: 320, top: 252, right: 432, bottom: 292}, "Cancel", color(45, 52, 66))
}

func paintAbout(hWnd uintptr) {
	var ps paintStruct
	hdc, _, _ := procBeginPaint.Call(hWnd, uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(hWnd, uintptr(unsafe.Pointer(&ps)))

	drawRound(hdc, rect{left: 0, top: 0, right: 560, bottom: 410}, color(15, 17, 23), color(15, 17, 23), 0)
	drawText(hdc, appName, rect{left: 28, top: 24, right: 520, bottom: 58}, titleFont, color(239, 244, 255), dtLeft|dtSingleLine)
	drawText(hdc, "About it", rect{left: 28, top: 60, right: 520, bottom: 86}, uiFont, color(150, 160, 178), dtLeft|dtSingleLine)
	drawRound(hdc, rect{left: 28, top: 104, right: 528, bottom: 278}, color(24, 28, 38), color(55, 65, 81), 16)
	drawText(hdc, aboutProjectText, rect{left: 48, top: 124, right: 508, bottom: 258}, uiFont, color(226, 232, 240), dtLeft|dtWordBreak)
	drawText(hdc, aboutWatermark, rect{left: 28, top: 296, right: 390, bottom: 326}, uiFont, color(150, 160, 178), dtLeft|dtSingleLine)
	drawButton(hdc, rect{left: 410, top: 300, right: 506, bottom: 338}, "OK", color(37, 99, 235))
}

func handleSettingsClick(hWnd uintptr, pt point) {
	switch {
	case inRect(pt, rect{left: 28, top: 102, right: 190, bottom: 166}):
		timeEdit, _, _ := procGetDlgItem.Call(hWnd, timeEditID)
		procSetFocus.Call(timeEdit)
	case inRect(pt, rect{left: 28, top: 190, right: 432, bottom: 218}):
		settingsDraftShutdown = !settingsDraftShutdown
	case inRect(pt, rect{left: 208, top: 252, right: 306, bottom: 292}):
		if applySettings(hWnd) {
			procDestroyWindow.Call(hWnd)
			return
		}
	case inRect(pt, rect{left: 320, top: 252, right: 432, bottom: 292}):
		procDestroyWindow.Call(hWnd)
		return
	}
	procInvalidateRect.Call(hWnd, 0, 1)
}

func showMenu() {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)

	timerMenu, _, _ := procCreatePopupMenu.Call()
	if timerMenu != 0 {
		for index, quickTimer := range quickTimers {
			procAppendMenuW.Call(timerMenu, mfString, uintptr(quickTimerCommandBase+index), ptr(quickTimer.label))
		}
	}

	procAppendMenuW.Call(menu, mfString, turnOnCommand, ptr("Turn on permanent"))
	procAppendMenuW.Call(menu, mfString, turnOffCommand, ptr("Turn off"))
	if timerMenu != 0 {
		procAppendMenuW.Call(menu, mfPopup, timerMenu, ptr("Timer"))
	}
	procAppendMenuW.Call(menu, mfSeparator, 0, 0)
	procAppendMenuW.Call(menu, mfString, settingsCommand, ptr("Settings"))
	procAppendMenuW.Call(menu, mfString, aboutCommand, ptr("About it"))
	procAppendMenuW.Call(menu, mfString, exitCommand, ptr("Exit"))

	var cursor point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&cursor)))
	procSetForegroundWindow.Call(mainWindow)

	command, _, _ := procTrackPopupMenu.Call(
		menu,
		tpmRightButton|tpmReturnCmd,
		uintptr(cursor.x),
		uintptr(cursor.y),
		0,
		mainWindow,
		0,
	)

	handleCommand(command)
}

func handleCommand(command uintptr) {
	switch command {
	case turnOnCommand:
		turnOnPermanent()
	case turnOffCommand:
		turnOff()
	case settingsCommand:
		openSettings()
	case aboutCommand:
		openAbout()
	case exitCommand:
		procDestroyWindow.Call(mainWindow)
	default:
		if command >= quickTimerCommandBase && command < quickTimerCommandBase+uintptr(len(quickTimers)) {
			quickTimer := quickTimers[int(command-quickTimerCommandBase)]
			turnOnFor(time.Duration(quickTimer.minutes) * time.Minute)
		}
	}
}

func openSettings() {
	if isLiveWindow(settingsWindow) {
		procSetForegroundWindow.Call(settingsWindow)
		return
	}
	settingsWindow = 0
	go runSettingsWindow()
}

func runSettingsWindow() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	settingsWindow, _, _ = procCreateWindowExW.Call(
		0,
		ptr(appName+"SettingsWindow"),
		ptr(appName+" Settings"),
		wsCaption|wsSysMenu|wsVisible,
		uintptr(scale(420)), uintptr(scale(260)), uintptr(scale(460)), uintptr(scale(330)),
		0, 0, hInstance, 0,
	)
	showTopWindow(settingsWindow, "Settings could not be opened.")
	runMessageLoop()
}

func openAbout() {
	if isLiveWindow(aboutWindow) {
		procSetForegroundWindow.Call(aboutWindow)
		return
	}
	aboutWindow = 0
	go runAboutWindow()
}

func runAboutWindow() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	aboutWindow, _, _ = procCreateWindowExW.Call(
		0,
		ptr(appName+"AboutWindow"),
		ptr("About "+appName),
		wsCaption|wsSysMenu|wsVisible,
		uintptr(scale(420)), uintptr(scale(260)), uintptr(scale(590)), uintptr(scale(450)),
		0, 0, hInstance, 0,
	)
	showTopWindow(aboutWindow, "About could not be opened.")
	runMessageLoop()
}

func runMessageLoop() {
	var message msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&message)))
	}
}

func isLiveWindow(hWnd uintptr) bool {
	if hWnd == 0 {
		return false
	}
	ret, _, _ := procIsWindow.Call(hWnd)
	return ret != 0
}

func showTopWindow(hWnd uintptr, errorText string) {
	if hWnd == 0 {
		procMessageBoxW.Call(mainWindow, ptr(errorText), ptr(appName), mbIconError)
		return
	}
	procShowWindow.Call(hWnd, swShow)
	procUpdateWindow.Call(hWnd)
	procSetForegroundWindow.Call(hWnd)
}

func turnOnPermanent() {
	active = true
	activeUntil = time.Time{}
	shutdownWhenOff = false
	syncPowerState()
	updateTrayIcon()
	sendF24()
}

func turnOff() {
	active = false
	activeUntil = time.Time{}
	shutdownWhenOff = false
	syncPowerState()
	updateTrayIcon()
}

func turnOnFor(duration time.Duration) {
	active = true
	activeUntil = time.Now().Add(duration)
	shutdownWhenOff = false
	syncPowerState()
	updateTrayIcon()
	sendF24()
}

func tick() {
	if !active {
		syncPowerState()
		return
	}
	if !activeUntil.IsZero() && time.Now().After(activeUntil) {
		shouldShutdown := shutdownWhenOff
		turnOff()
		if shouldShutdown {
			shutdownPC()
		}
		return
	}
	keepSystemAwake()
	sendF24()
}

func syncPowerState() {
	if active {
		keepSystemAwake()
		return
	}
	allowSystemSleep()
}

func keepSystemAwake() {
	procSetThreadExecutionState.Call(esContinuous | esSystemRequired)
}

func allowSystemSleep() {
	procSetThreadExecutionState.Call(esContinuous)
}

func notifyPowerStateChanged() {
	if mainWindow != 0 {
		procPostMessageW.Call(mainWindow, wmPowerUpdate, 0, 0)
	}
}

func addTrayIcon() {
	data := trayData(nimAdd)
	procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&data)))
}

func updateTrayIcon() {
	data := trayData(nimModify)
	procShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&data)))
}

func removeTrayIcon() {
	data := notifyIconData{
		cbSize: uint32(unsafe.Sizeof(notifyIconData{})),
		hWnd:   mainWindow,
		uID:    1,
	}
	procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&data)))
}

func trayData(action uint32) notifyIconData {
	_ = action
	icon := inactiveIcon
	tip := appName + " off"
	if active {
		icon = activeIcon
		tip = appName + " on"
		if !activeUntil.IsZero() {
			tip = appName + " on until " + activeUntil.Format("15:04")
		}
	}

	data := notifyIconData{
		cbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		hWnd:             mainWindow,
		uID:              1,
		uFlags:           nifMessage | nifIcon | nifTip,
		uCallbackMessage: trayMessage,
		hIcon:            icon,
	}
	copy(data.szTip[:], syscall.StringToUTF16(tip))
	return data
}

func sendF24() {
	events := [2]input{
		{
			inputType: inputKeyboard,
			ki: keyboardInput{
				wVk: vkF24,
			},
		},
		{
			inputType: inputKeyboard,
			ki: keyboardInput{
				wVk:     vkF24,
				dwFlags: keyEventFKeyUp,
			},
		},
	}

	procSendInput.Call(
		uintptr(len(events)),
		uintptr(unsafe.Pointer(&events[0])),
		unsafe.Sizeof(events[0]),
	)
}

func shutdownPC() {
	_ = exec.Command("shutdown.exe", "/s", "/t", "0").Start()
}

func readWindowText(hWnd uintptr) string {
	buffer := make([]uint16, 64)
	procGetWindowTextW.Call(hWnd, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	return syscall.UTF16ToString(buffer)
}

func iconFromICO(ico []byte) uintptr {
	if len(ico) < 22 || binary.LittleEndian.Uint16(ico[0:2]) != 0 || binary.LittleEndian.Uint16(ico[2:4]) != 1 {
		return 0
	}

	count := int(binary.LittleEndian.Uint16(ico[4:6]))
	for _, preferredSize := range []int{32, 48, 64, 256, 24, 16} {
		for i := 0; i < count; i++ {
			entry := 6 + i*16
			if len(ico) < entry+16 {
				break
			}
			width := int(ico[entry])
			if width == 0 {
				width = 256
			}
			if width != preferredSize {
				continue
			}

			if icon := iconFromICOEntry(ico, entry); icon != 0 {
				return icon
			}
		}
	}

	for i := 0; i < count; i++ {
		entry := 6 + i*16
		if len(ico) < entry+16 {
			break
		}
		if icon := iconFromICOEntry(ico, entry); icon != 0 {
			return icon
		}
	}
	return 0
}

func iconFromICOEntry(ico []byte, entry int) uintptr {
	size := int(binary.LittleEndian.Uint32(ico[entry+8 : entry+12]))
	offset := int(binary.LittleEndian.Uint32(ico[entry+12 : entry+16]))
	if offset < 0 || size <= 0 || offset+size > len(ico) {
		return 0
	}

	icon, _, _ := procCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&ico[offset])),
		uintptr(size),
		1,
		0x00030000,
		32,
		32,
		0,
	)
	return icon
}

func lowWord(value uintptr) uintptr {
	return value & 0xffff
}

func ptr(value string) uintptr {
	return uintptr(unsafe.Pointer(utf16Ptr(value)))
}

func utf16Ptr(value string) *uint16 {
	ptr, err := syscall.UTF16PtrFromString(value)
	if err != nil {
		return nil
	}
	return ptr
}
