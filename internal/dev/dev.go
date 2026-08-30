package dev

import (
	"fmt"
	"strings"
	"github.com/out-lang/out/internal/module"
	"github.com/out-lang/out/internal/object"
)

type Board struct {
	Name        string
	DigitalPins int
	AnalogPins  int
	PWMPins     []int
	Voltage     float64
	FlashKB     int
	SRAMKB      int
}

var boards = map[string]*Board{
	"arduino nano": {
		Name:        "Arduino Nano",
		DigitalPins: 14,
		AnalogPins:  6,
		PWMPins:     []int{3, 5, 6, 9, 10, 11},
		Voltage:     5.0,
		FlashKB:     32,
		SRAMKB:      2,
	},
	"arduino uno": {
		Name:        "Arduino Uno",
		DigitalPins: 14,
		AnalogPins:  6,
		PWMPins:     []int{3, 5, 6, 9, 10, 11},
		Voltage:     5.0,
		FlashKB:     32,
		SRAMKB:      2,
	},
	"arduino mega": {
		Name:        "Arduino Mega 2560",
		DigitalPins: 54,
		AnalogPins:  16,
		PWMPins:     []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13},
		Voltage:     5.0,
		FlashKB:     256,
		SRAMKB:      8,
	},
	"arduino leonardo": {
		Name:        "Arduino Leonardo",
		DigitalPins: 14,
		AnalogPins:  6,
		PWMPins:     []int{3, 5, 6, 9, 10, 11, 13},
		Voltage:     5.0,
		FlashKB:     32,
		SRAMKB:      3,
	},
	"arduino due": {
		Name:        "Arduino Due",
		DigitalPins: 54,
		AnalogPins:  12,
		PWMPins:     []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13},
		Voltage:     3.3,
		FlashKB:     512,
		SRAMKB:      96,
	},
	"esp32": {
		Name:        "ESP32",
		DigitalPins: 40,
		AnalogPins:  16,
		PWMPins:     []int{2, 4, 5, 12, 13, 14, 15, 16, 17, 18, 19, 21, 22, 23, 25, 26, 27, 32, 33},
		Voltage:     3.3,
		FlashKB:     4096,
		SRAMKB:      520,
	},
	"esp32s2": {
		Name:        "ESP32-S2",
		DigitalPins: 43,
		AnalogPins:  20,
		PWMPins:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18},
		Voltage:     3.3,
		FlashKB:     4096,
		SRAMKB:      320,
	},
	"esp32s3": {
		Name:        "ESP32-S3",
		DigitalPins: 45,
		AnalogPins:  20,
		PWMPins:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 21, 38, 39, 40, 42},
		Voltage:     3.3,
		FlashKB:     8192,
		SRAMKB:      512,
	},
	"esp32c3": {
		Name:        "ESP32-C3",
		DigitalPins: 22,
		AnalogPins:  7,
		PWMPins:     []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Voltage:     3.3,
		FlashKB:     4096,
		SRAMKB:      400,
	},
	"esp8266": {
		Name:        "ESP8266 (NodeMCU)",
		DigitalPins: 17,
		AnalogPins:  1,
		PWMPins:     []int{0, 1, 2, 3, 4, 5, 12, 13, 14, 15, 16},
		Voltage:     3.3,
		FlashKB:     4096,
		SRAMKB:      80,
	},
	"esp01": {
		Name:        "ESP-01",
		DigitalPins: 4,
		AnalogPins:  0,
		PWMPins:     []int{2},
		Voltage:     3.3,
		FlashKB:     512,
		SRAMKB:      36,
	},
	"esp01s": {
		Name:        "ESP-01S",
		DigitalPins: 4,
		AnalogPins:  0,
		PWMPins:     []int{2},
		Voltage:     3.3,
		FlashKB:     1024,
		SRAMKB:      36,
	},
	"raspberry pi pico": {
		Name:        "Raspberry Pi Pico",
		DigitalPins: 30,
		AnalogPins:  3,
		PWMPins:     []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27},
		Voltage:     3.3,
		FlashKB:     2048,
		SRAMKB:      264,
	},
	"stm32f103": {
		Name:        "STM32 Blue Pill",
		DigitalPins: 37,
		AnalogPins:  10,
		PWMPins:     []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		Voltage:     3.3,
		FlashKB:     64,
		SRAMKB:      20,
	},
	"stm32f407": {
		Name:        "STM32 Discovery",
		DigitalPins: 42,
		AnalogPins:  16,
		PWMPins:     []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		Voltage:     3.3,
		FlashKB:     1024,
		SRAMKB:      192,
	},
	"attiny85": {
		Name:        "ATtiny85",
		DigitalPins: 6,
		AnalogPins:  4,
		PWMPins:     []int{0, 1, 4},
		Voltage:     5.0,
		FlashKB:     8,
		SRAMKB:      1,
	},
	"atmega328p": {
		Name:        "ATmega328P (bare)",
		DigitalPins: 23,
		AnalogPins:  8,
		PWMPins:     []int{3, 5, 6, 9, 10, 11},
		Voltage:     5.0,
		FlashKB:     32,
		SRAMKB:      2,
	},
	"teensy 4.0": {
		Name:        "Teensy 4.0",
		DigitalPins: 40,
		AnalogPins:  14,
		PWMPins:     []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 22, 23, 24, 25, 28, 29, 33, 34, 35, 36, 37, 38, 39},
		Voltage:     3.3,
		FlashKB:     2048,
		SRAMKB:      1024,
	},
	"teensy 3.2": {
		Name:        "Teensy 3.2",
		DigitalPins: 34,
		AnalogPins:  12,
		PWMPins:     []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 20, 21, 22, 23},
		Voltage:     3.3,
		FlashKB:     256,
		SRAMKB:      64,
	},
	"nodemcu": {
		Name:        "NodeMCU v3",
		DigitalPins: 17,
		AnalogPins:  1,
		PWMPins:     []int{0, 1, 2, 3, 4, 5, 12, 13, 14, 15, 16},
		Voltage:     3.3,
		FlashKB:     4096,
		SRAMKB:      80,
	},
	"wemos d1 mini": {
		Name:        "Wemos D1 Mini",
		DigitalPins: 17,
		AnalogPins:  1,
		PWMPins:     []int{0, 1, 2, 3, 4, 5, 12, 13, 14, 15, 16},
		Voltage:     3.3,
		FlashKB:     4096,
		SRAMKB:      80,
	},
	"feather esp32s3": {
		Name:        "Adafruit Feather ESP32-S3",
		DigitalPins: 45,
		AnalogPins:  20,
		PWMPins:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 21, 38, 39, 40, 42},
		Voltage:     3.3,
		FlashKB:     8192,
		SRAMKB:      512,
	},
	"xiao esp32c3": {
		Name:        "Seeed XIAO ESP32-C3",
		DigitalPins: 11,
		AnalogPins:  5,
		PWMPins:     []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Voltage:     3.3,
		FlashKB:     4096,
		SRAMKB:      400,
	},
	"rp2040": {
		Name:        "RP2040 (generic)",
		DigitalPins: 30,
		AnalogPins:  4,
		PWMPins:     []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27},
		Voltage:     3.3,
		FlashKB:     2048,
		SRAMKB:      264,
	},
}

var (
	currentBoard *Board
	pins         = make(map[int64]int64)
	pinModes     = make(map[int64]string)
)

func validateDigitalPin(pin int64) string {
	if currentBoard == nil {
		return ""
	}
	if pin < 0 || pin >= int64(currentBoard.DigitalPins) {
		return fmt.Sprintf("pin %d does not exist on %s (digital pins: 0-%d)", pin, currentBoard.Name, currentBoard.DigitalPins-1)
	}
	return ""
}

func validateAnalogPin(pin int64) string {
	if currentBoard == nil {
		return ""
	}
	if pin < 0 || pin >= int64(currentBoard.AnalogPins) {
		return fmt.Sprintf("analog pin A%d does not exist on %s (analog pins: A0-A%d)", pin, currentBoard.Name, currentBoard.AnalogPins-1)
	}
	return ""
}

func isPWMPin(pin int64) bool {
	if currentBoard == nil {
		return true
	}
	for _, p := range currentBoard.PWMPins {
		if int64(p) == pin {
			return true
		}
	}
	return false
}

func Module() *module.Module {
	m := module.New("dev")

	m.Set("board", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "dev::board expects 1 argument (board name)"}
		}
		name, ok := args[0].(*object.String)
		if !ok {
			return &object.Error{Message: "dev::board expects STRING"}
		}
		key := strings.ToLower(name.Value)
		b, found := boards[key]
		if !found {
			known := make([]string, 0, len(boards))
			for k := range boards {
				known = append(known, k)
			}
			return &object.Error{Message: fmt.Sprintf("unknown board '%s'. available: %s", name.Value, strings.Join(known, ", "))}
		}
		currentBoard = b
		pins = make(map[int64]int64)
		pinModes = make(map[int64]string)
		fmt.Printf("[dev] board: %s (%d digital, %d analog, %dV, %dKB flash)\n",
			b.Name, b.DigitalPins, b.AnalogPins, int64(b.Voltage), b.FlashKB)
		return &object.Null{}
	})

	m.Set("info", func(args ...object.Object) object.Object {
		if currentBoard == nil {
			return &object.String{Value: "no board configured. use dev::board(\"arduino nano\") first"}
		}
		b := currentBoard
		pwmStr := make([]string, len(b.PWMPins))
		for i, p := range b.PWMPins {
			pwmStr[i] = fmt.Sprintf("D%d", p)
		}
		info := fmt.Sprintf("Board: %s | Digital: %d (D0-D%d) | Analog: %d (A0-A%d) | PWM: %s | Flash: %dKB | SRAM: %dKB | Voltage: %dV",
			b.Name, b.DigitalPins, b.DigitalPins-1,
			b.AnalogPins, b.AnalogPins-1,
			strings.Join(pwmStr, ", "),
			b.FlashKB, b.SRAMKB, int64(b.Voltage))
		return &object.String{Value: info}
	})

	m.Set("boards", func(args ...object.Object) object.Object {
		names := make([]object.Object, 0, len(boards))
		for _, b := range boards {
			names = append(names, &object.String{Value: b.Name})
		}
		return &object.Array{Elements: names}
	})

	m.Set("pinMode", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "dev::pinMode expects 2 arguments (pin, mode)"}
		}
		pin, ok := args[0].(*object.Integer)
		if !ok {
			return &object.Error{Message: "dev::pinMode expects INTEGER for pin"}
		}
		mode, ok := args[1].(*object.String)
		if !ok {
			return &object.Error{Message: "dev::pinMode expects STRING for mode (INPUT, OUTPUT, INPUT_PULLUP)"}
		}
		if err := validateDigitalPin(pin.Value); err != "" {
			return &object.Error{Message: err}
		}
		validModes := map[string]bool{"INPUT": true, "OUTPUT": true, "INPUT_PULLUP": true}
		if !validModes[mode.Value] {
			return &object.Error{Message: fmt.Sprintf("invalid mode '%s'. use INPUT, OUTPUT, or INPUT_PULLUP", mode.Value)}
		}
		pinModes[pin.Value] = mode.Value
		fmt.Printf("[dev] pinMode(D%d, %s) on %s\n", pin.Value, mode.Value, currentBoard.Name)
		return &object.Null{}
	})

	m.Set("digitalWrite", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "dev::digitalWrite expects 2 arguments (pin, value)"}
		}
		pin, ok := args[0].(*object.Integer)
		if !ok {
			return &object.Error{Message: "dev::digitalWrite expects INTEGER for pin"}
		}
		val, ok := args[1].(*object.Integer)
		if !ok {
			return &object.Error{Message: "dev::digitalWrite expects INTEGER for value (0/1)"}
		}
		if err := validateDigitalPin(pin.Value); err != "" {
			return &object.Error{Message: err}
		}
		if val.Value != 0 && val.Value != 1 {
			return &object.Error{Message: "digitalWrite value must be 0 (LOW) or 1 (HIGH)"}
		}
		pins[pin.Value] = val.Value
		pwm := ""
		if currentBoard != nil && isPWMPin(pin.Value) {
			pwm = " [PWM]"
		}
		fmt.Printf("[dev] digitalWrite(D%d, %d)%s on %s\n", pin.Value, val.Value, pwm, currentBoard.Name)
		return &object.Null{}
	})

	m.Set("digitalRead", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "dev::digitalRead expects 1 argument (pin)"}
		}
		pin, ok := args[0].(*object.Integer)
		if !ok {
			return &object.Error{Message: "dev::digitalRead expects INTEGER for pin"}
		}
		if err := validateDigitalPin(pin.Value); err != "" {
			return &object.Error{Message: err}
		}
		val := pins[pin.Value]
		fmt.Printf("[dev] digitalRead(D%d) -> %d on %s\n", pin.Value, val, currentBoard.Name)
		return &object.Integer{Value: val}
	})

	m.Set("analogRead", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "dev::analogRead expects 1 argument (analog pin)"}
		}
		pin, ok := args[0].(*object.Integer)
		if !ok {
			return &object.Error{Message: "dev::analogRead expects INTEGER for analog pin (0-5 on Nano)"}
		}
		if err := validateAnalogPin(pin.Value); err != "" {
			return &object.Error{Message: err}
		}
		val := pins[pin.Value]
		fmt.Printf("[dev] analogRead(A%d) -> %d on %s\n", pin.Value, val, currentBoard.Name)
		return &object.Integer{Value: val}
	})

	m.Set("analogWrite", func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "dev::analogWrite expects 2 arguments (pin, value)"}
		}
		pin, ok := args[0].(*object.Integer)
		if !ok {
			return &object.Error{Message: "dev::analogWrite expects INTEGER for pin"}
		}
		val, ok := args[1].(*object.Integer)
		if !ok {
			return &object.Error{Message: "dev::analogWrite expects INTEGER for value (0-255)"}
		}
		if err := validateDigitalPin(pin.Value); err != "" {
			return &object.Error{Message: err}
		}
		if val.Value < 0 || val.Value > 255 {
			return &object.Error{Message: "analogWrite value must be 0-255"}
		}
		if currentBoard != nil && !isPWMPin(pin.Value) {
			return &object.Error{Message: fmt.Sprintf("pin D%d is not PWM on %s. PWM pins: %v", pin.Value, currentBoard.Name, currentBoard.PWMPins)}
		}
		pins[pin.Value] = val.Value
		fmt.Printf("[dev] analogWrite(D%d, %d) on %s\n", pin.Value, val.Value, currentBoard.Name)
		return &object.Null{}
	})

	m.Set("delay", func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "dev::delay expects 1 argument (milliseconds)"}
		}
		ms, ok := args[0].(*object.Integer)
		if !ok {
			return &object.Error{Message: "dev::delay expects INTEGER (milliseconds)"}
		}
		fmt.Printf("[dev] delay(%dms)\n", ms.Value)
		return &object.Null{}
	})

	m.Set("touch_grass", func(args ...object.Object) object.Object {
		if currentBoard != nil {
			fmt.Printf("[dev] touch_grass() -> go outside, disconnect %s, and touch grass\n", currentBoard.Name)
		} else {
			fmt.Println("[dev] touch_grass() -> go outside and touch grass")
		}
		return &object.String{Value: "grass touched. you feel better now."}
	})

	m.Desc = "Hardware dev: board selection, pin control, analog/digital I/O"
	return m
}
