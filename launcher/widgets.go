package launcher

import (
	"fmt"
	"github.com/AllenDang/cimgui-go"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"golang.org/x/exp/constraints"
	"strconv"
	"strings"
)

type popupType int

const (
	popDynamic = popupType(iota)
	popMedium
	popBig
)

var sliderSledLastFrame = false
var sliderSledThisFrame = false

type iPopup interface {
	draw()
	shouldClose() bool
	open()
}

type popup struct {
	internalDraw func()

	name string

	popType popupType

	opened bool

	closeListener  func()
	listenerCalled bool
}

func newPopup(name string, popType popupType) *popup {
	return &popup{
		name:    name,
		popType: popType,
	}
}

func newPopupF(name string, popType popupType, draw func()) *popup {
	return &popup{
		name:         name,
		popType:      popType,
		internalDraw: draw,
	}
}

func (p *popup) draw() {
	p.opened = true
	switch p.popType {
	case popDynamic, popMedium:
		popupSmall(p.name, &p.opened, p.popType == popDynamic, p.internalDraw)
	case popBig:
		popupWide(p.name, &p.opened, p.internalDraw)
	}
}

func (p *popup) shouldClose() bool {
	if !p.opened && !p.listenerCalled {
		if p.closeListener != nil {
			p.closeListener()
		}

		p.listenerCalled = true
	}

	return !p.opened
}

func (p *popup) open() {
	p.listenerCalled = false
	p.opened = true
}

func (p *popup) setCloseListener(closeListener func()) {
	p.closeListener = closeListener
}

var openedAbove bool

func resetPopupHierarchyInfo() {
	openedAbove = false
}

func popupSmall(name string, opened *bool, dynamicSize bool, content func()) {
	width := float32(settings.Graphics.WindowWidth)

	sX := width / 2
	if dynamicSize {
		sX = 0
	}

	popupInter(name, opened, imgui.Vec2{X: sX, Y: 0}, content)
}

func popupWide(name string, opened *bool, content func()) {
	wSize := imgui.WindowSize()
	popupInter(name, opened, imgui.Vec2{X: wSize.X * 0.9, Y: wSize.Y * 0.9}, content)
}

func popupInter(name string, opened *bool, size imgui.Vec2, content func()) {
	wSizeX, wSizeY := float32(settings.Graphics.WindowWidth), float32(settings.Graphics.WindowHeight)

	if *opened {
		if !imgui.IsPopupOpenStr("##" + name) {
			imgui.OpenPopupStr("##" + name)
		}

		imgui.SetNextWindowSize(size)

		imgui.SetNextWindowPosV(imgui.Vec2{
			X: wSizeX / 2,
			Y: wSizeY / 2,
		}, imgui.CondAlways, imgui.Vec2{
			X: 0.5,
			Y: 0.5,
		})

		if imgui.BeginPopupModalV("##"+name, opened, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoResize|imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar) {
			content()

			hovered := imgui.IsWindowHoveredV(imgui.HoveredFlagsRootAndChildWindows|imgui.HoveredFlagsAllowWhenBlockedByActiveItem|imgui.HoveredFlagsAllowWhenBlockedByPopup) || openedAbove

			if ((imgui.IsMouseClickedBool(0) || imgui.IsMouseClickedBool(1)) && !hovered) || imgui.IsKeyPressedBool(imgui.KeyEscape) {
				*opened = false
			}

			openedAbove = true

			imgui.EndPopup()
		}
	}
}

func sliderFloatReset(label string, val *floatParam, min, max float32, format string) {
	sliderFloatResetBase(label, val, min, max, func() bool {
		return sliderFloatSlide("##"+label, &val.value, min, max, format, imgui.SliderFlagsNoInput)
	})
}

func sliderFloatResetStep(label string, val *floatParam, min, max, step float32, format string) {
	sliderFloatResetBase(label, val, min, max, func() bool {
		return sliderFloatStep("##"+label, &val.value, min, max, step, format)
	})
}

func sliderFloatResetBase(label string, val *floatParam, min, max float32, sliderFunc func() bool) {
	if val.value < min || val.value > max {
		val.value = mutils.Clamp(val.value, min, max)
		paramChanged(val)
	}

	sliderResetBase(label, !val.changed, func() {
		if sliderFunc() {
			paramChanged(val)
		}
	}, func() {
		val.value = val.ogValue
		val.changed = false
	})
}

func paramChanged(val *floatParam) {
	if math32.Abs(val.value-val.ogValue) > 0.001 {
		val.changed = true
	} else {
		val.changed = false
		val.value = val.ogValue
	}
}

func sliderIntReset(label string, val *intParam, min, max int32, format string) {
	sliderResetBase(label, !val.changed, func() {
		if sliderIntSlide("##"+label, &val.value, min, max, format, imgui.SliderFlagsNoInput) {
			val.changed = val.value != val.ogValue
		}
	}, func() {
		val.value = val.ogValue
		val.changed = false
	})
}

func sliderResetBase(label string, blockButton bool, draw, reset func()) {
	imgui.Text(label + ":")

	imgui.PushFont(Font16)

	if imgui.BeginTableV("rt"+label, 2, imgui.TableFlagsSizingStretchProp, vec2(-1, 0), -1) {
		imgui.TableSetupColumnV("rt1"+label, imgui.TableColumnFlagsWidthStretch, 0, imgui.ID(0))
		imgui.TableSetupColumnV("rt2"+label, imgui.TableColumnFlagsWidthFixed, 0, imgui.ID(1))

		imgui.TableNextColumn()

		imgui.SetNextItemWidth(-1)

		draw()

		imgui.TableNextColumn()

		if blockButton {
			imgui.BeginDisabled()
		}

		if imgui.Button("Reset##" + label) {
			reset()
		}

		if blockButton {
			imgui.EndDisabled()
		}

		imgui.EndTable()
	}

	imgui.PopFont()
}

func sliderFloatStep(label string, val *float32, min, max, step float32, format string) bool {
	stepNum := int32((max - min) / step)

	v := int32(math32.Round((*val - min) / step))

	cPos := imgui.CursorPos()
	iW := imgui.CalcItemWidth() + imgui.CurrentStyle().FramePadding().X*2

	ret := sliderIntSlide(label, &v, 0, stepNum, "##%d", imgui.SliderFlagsNoInput)

	cPos2 := imgui.CursorPos()

	*val = (float32(v) * step) + min

	tx := fmt.Sprintf(format, *val)

	tS := imgui.CalcTextSizeV(tx+"f", false, 0)

	imgui.SetCursorPos(imgui.Vec2{
		X: cPos.X + (iW-tS.X)/2,
		Y: cPos.Y,
	})

	imgui.AlignTextToFramePadding()

	imgui.Text(tx)

	imgui.SetCursorPos(cPos2)
	imgui.Dummy(vzero())

	return ret
}

func sliderIntSlide(label string, value *int32, min, max int32, format string, flags imgui.SliderFlags) (ret bool) {
	ret = imgui.SliderIntV(label, value, min, max, format, flags)

	sliderSledThisFrame = sliderSledThisFrame || ret

	if imgui.IsItemHovered() || imgui.IsItemActive() {
		ret = ret || keySlideInt(value, min, max)
	}

	return
}

func sliderFloatSlide(label string, value *float32, min, max float32, format string, flags imgui.SliderFlags) (ret bool) {
	ret = imgui.SliderFloatV(label, value, min, max, format, flags)

	sliderSledThisFrame = sliderSledThisFrame || ret

	if imgui.IsItemHovered() || imgui.IsItemActive() {
		iDot := strings.Index(format, ".")
		iF := strings.Index(format, "f")
		prec, _ := strconv.Atoi(format[iDot+1 : iF])
		step := math32.Pow(10, -float32(prec))

		ret = ret || keySlideFloat(value, min, max, step)
	}

	return
}

func keySlideInt[T constraints.Integer](value *T, min, max T) (ret bool) {
	if imgui.IsKeyPressedBool(imgui.KeyLeftArrow) {
		*value = mutils.Clamp(*value-1, min, max)
		ret = true
	}

	if imgui.IsKeyPressedBool(imgui.KeyRightArrow) {
		*value = mutils.Clamp(*value+1, min, max)
		ret = true
	}

	return
}

func keySlideFloat[T constraints.Float](value *T, min, max, step T) (ret bool) {
	if imgui.IsKeyPressedBool(imgui.KeyLeftArrow) {
		*value = mutils.Clamp(*value-step, min, max)
		ret = true
	}

	if imgui.IsKeyPressedBool(imgui.KeyRightArrow) {
		*value = mutils.Clamp(*value+step, min, max)
		ret = true
	}

	return
}

func centerTable(label string, width float32, draw func()) {
	if imgui.BeginTableV(label, 3, imgui.TableFlagsSizingStretchProp, vec2(width, 0), -1) {
		imgui.TableSetupColumnV("1"+label, imgui.TableColumnFlagsWidthStretch, 0, imgui.ID(0))
		imgui.TableSetupColumnV("2"+label, imgui.TableColumnFlagsWidthFixed, 0, imgui.ID(1))
		imgui.TableSetupColumnV("3"+label, imgui.TableColumnFlagsWidthStretch, 0, imgui.ID(2))

		imgui.TableNextColumn()

		imgui.Dummy(vec2(1, 1))

		imgui.TableNextColumn()

		draw()

		imgui.TableNextColumn()

		imgui.Dummy(vec2(1, 1))

		imgui.EndTable()
	}
}

func selectableFocus(label string, selected, justOpened bool) (clicked bool) {
	if selected && justOpened {
		imgui.SetScrollYFloat(imgui.CursorPosY()) //SetScrollHereY was not working reliably
	}

	clicked = imgui.SelectableBoolV(label, selected, 0, imgui.Vec2{})

	if clicked {
		imgui.CloseCurrentPopup()
	}

	return
}

func searchBox(label string, searchString *string) (ok bool) {
	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 0)
	imgui.PushStyleVarFloat(imgui.StyleVarFrameBorderSize, 0)

	imgui.PushStyleColorVec4(imgui.ColFrameBg, vec4(0, 0, 0, 1))
	imgui.PushStyleColorVec4(imgui.ColFrameBgActive, vec4(0.1, 0.1, 0.1, 1))
	imgui.PushStyleColorVec4(imgui.ColFrameBgHovered, vec4(0.1, 0.1, 0.1, 1))

	ok = imgui.InputTextWithHint(label, "Search", searchString, imgui.InputTextFlagsNone, nil)

	imgui.PopStyleColor()
	imgui.PopStyleColor()
	imgui.PopStyleColor()

	imgui.PopStyleVar()
	imgui.PopStyleVar()

	return
}

func inputText(label string, text *string) bool {
	return inputTextV(label, text, imgui.InputTextFlagsNone, nil)
}

func inputTextV(label string, text *string, flags int, cb imgui.InputTextCallback) bool {
	return imgui.InputTextWithHint(label, "", text, imgui.InputTextFlags(flags), cb)
}
