package doser

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/reef-pi/reef-pi/controller/connectors"
	"github.com/reef-pi/reef-pi/controller/device_manager"
	"github.com/reef-pi/reef-pi/controller/telemetry"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	gpio2 "periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/host"
)

type Runner struct {
	deviceManager *device_manager.DeviceManager
	devMode       bool
	pump          *Pump
	jacks         *connectors.Jacks
	statsMgr      telemetry.StatsManager
}

func (runner *Runner) DoseStepper(speed float64, duration float64) {
	//no need to dose in the devMode
	if runner.devMode {
		return
	}

	initHertz, err := strconv.Atoi(runner.pump.In1Pin)
	if err != nil {
		log.Fatal(err)
	}
	operationHertz, err := strconv.Atoi(runner.pump.In2Pin)
	if err != nil {
		log.Fatal(err)
	}

	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
	// Use gpioreg GPIO pin registry to find a GPIO pin by name.
	stepPin := gpioreg.ByName("GPIO24")
	dirPin := gpioreg.ByName("GPIO25")
	ms1Pin := gpioreg.ByName("GPIO23")
	ms2Pin := gpioreg.ByName("GPIO18")
	if stepPin == nil || dirPin == nil {
		log.Fatal("Failed to find GPIO24 or GPIO25")
	}

	if err := dirPin.Out(gpio2.High); err != nil {
		log.Fatal(err)
	}

	//set stepping to half steps --> ms1 = 1, ms2 = 0
	//https://www.electronicoscaldas.com/datasheet/A3967-EDMOD_Manual.pdf
	if err := ms1Pin.Out(gpio2.High); err != nil {
		log.Fatal(err)
	}
	if err := ms2Pin.Out(gpio2.Low); err != nil {
		log.Fatal(err)
	}
	if err := stepPin.PWM(gpio2.DutyHalf, physic.Frequency(initHertz)); err != nil {
		log.Fatal(err)
	}

	// when the pump needs to run longer, we should take care of vendor specific duty logic
	// in this case I am implementing the logic below to adhere to Welco's spec
	if duration > 3 {
		log.Println("init stepper motor at ", initHertz)
		time.Sleep(time.Duration(500 * time.Millisecond))

		log.Println("operate motor at ", operationHertz)
		if err := stepPin.PWM(gpio2.DutyHalf, physic.Frequency(operationHertz)); err != nil {
			log.Fatal(err)
		}
	}
	time.Sleep(time.Duration(duration * float64(time.Second)))

	//stop the stepper driver
	if err := stepPin.Out(gpio2.Low); err != nil {
		log.Fatal(err)
	}
	if err := dirPin.Out(gpio2.Low); err != nil {
		log.Fatal(err)
	}
}

func (runner *Runner) L298NDoseStepper(speed float64, duration float64) {

	//no need to dose in the devMode
	if runner.devMode {
		return
	}
	// Inputs to the DoseStepper stepDelay, seq => arrays of steps,
	// and stepDir; 1 or 2 for clockwise, -1 or -2 for counter-clockwise
	// Check stepDir and seq with manufacturer documentation
	// Start here: https://github.com/hybridgroup/gobot/blob/a8f33b2fc012951104857c485e85b35bf5c4cb9d/drivers/gpio/stepper_driver.go
	r := raspi.NewAdaptor()
	log.Printf("Stepper pins: %s, %s, %s, %s\n",
		runner.pump.In1Pin,
		runner.pump.In2Pin,
		runner.pump.In3Pin,
		runner.pump.In4Pin)

	stepper := gpio.NewStepperDriver(r,
		[4]string{runner.pump.In1Pin, //runner.pump.In1Pin,  //A1 -> red/brown    GPIO25 -> 22
			runner.pump.In2Pin,  //runner.pump.In2Pin,       //A2 -> yellow/black GPIO24 -> 18
			runner.pump.In3Pin,  //runner.pump.In3Pin,       //B1 -> orange/gray  GPIO14 -> 8
			runner.pump.In4Pin}, //runner.pump.In4Pin},      //B2 -> blue/white   GPIO23 -> 16
		gpio.StepperModes.SinglePhaseStepping,
		runner.pump.StepsPerRevolution)

	work := func() {
		//set spped
		stepper.SetSpeed(uint(speed))
		// stepper.SetDirection("forward")
		// maSpeedGpio.PwmWrite(maSpeed)
		//Move forward one revolution
		if err := stepper.Move(int(duration)); err != nil {
			fmt.Println(err)
		}
	}

	robot := gobot.NewRobot("stepperBot",
		[]gobot.Connection{r},
		[]gobot.Device{stepper},
		work,
	)

	robot.Start()
}
func (r *Runner) Dose(speed float64, duration float64) error {
	log.Println("In the DOSE function (speed, duration)", speed, duration)

	if r.pump.IsStepper {
		log.Printf("Stepper mode dosing speed:%v, duration:%v\n", speed, duration)
		//logic for stepper motor dosing
		r.DoseStepper(speed, duration)
	} else {
		// Make sure periph is initialized.
		if _, err := host.Init(); err != nil {
			log.Fatal(err)
		}
		// Use gpioreg GPIO pin registry to find a GPIO pin by name.
		pwmPin := gpioreg.ByName("GPIO7")
		dirPin := gpioreg.ByName("GPIO16")
		if pwmPin == nil || dirPin == nil {
			log.Fatal("Failed to find GPIO7 or GPIO16")
		}
		log.Println("PWM motor starting at speed of ", speed)
		if err := dirPin.Out(gpio2.Low); err != nil {
			log.Fatal(err)
		}

		if err := pwmPin.Out(gpio2.High); err != nil {
			log.Fatal(err)
		}
		speedToDuty := int(gpio2.DutyMax) * int(speed) / 100
		//start with 500hz as specified in the welco manual
		if err := pwmPin.PWM(gpio2.Duty(speedToDuty), 500*physic.Hertz); err != nil {
			log.Fatal(err)
		}

		// when the pump needs to run longer, we should take care of vendor specific duty logic
		// in this case I am implementing the logic below to adhere to Welco's spec
		if duration > 3 {
			log.Println("PWM motor at 500 hz")
			time.Sleep(time.Duration(100 * time.Millisecond))

			log.Println("PWM motor at 2500 hz duty")
			if err := pwmPin.PWM(gpio2.Duty(speedToDuty), 2500*physic.Hertz); err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Duration(100 * time.Millisecond))

			log.Println("PWM motor at 8000 hz duty")
			if err := pwmPin.PWM(gpio2.Duty(speedToDuty), 8000*physic.Hertz); err != nil {
				log.Fatal(err)
			}
		}
		time.Sleep(time.Duration(duration * float64(time.Second)))

		//reverse motor
		// if err := pwmPin.Out(gpio2.Low); err != nil {
		// 	log.Fatal(err)
		// }
		// if err := dirPin.Out(gpio2.Low); err != nil {
		// 	log.Fatal(err)
		// }
		// log.Println("PWM motor reversing at speed of ", speed)
		// if err := dirPin.Out(gpio2.High); err != nil {
		// 	log.Fatal(err)
		// }

		// if err := pwmPin.Out(gpio2.Low); err != nil {
		// 	log.Fatal(err)
		// }
		// if err := pwmPin.PWM(gpio2.Duty(0), 500*physic.Hertz); err != nil {
		// 	log.Fatal(err)
		// }
		// log.Println("PWM motor reverse at 500 hz")
		// time.Sleep(time.Duration(1 * time.Second))

		// log.Println("PWM motor at 8000 hz duty")
		// if err := pwmPin.PWM(gpio2.DutyMax, 8000*physic.Hertz); err != nil {
		// 	log.Fatal(err)
		// }
		// time.Sleep(time.Duration(duration * float64(time.Second)))

		if err := pwmPin.Out(gpio2.Low); err != nil {
			log.Fatal(err)
		}
		if err := dirPin.Out(gpio2.Low); err != nil {
			log.Fatal(err)
		}
		log.Println("PWM motor halted")

		// pinDir := rpio.Pin(16)
		// pinPWM := rpio.Pin(7)
		// pinPWM.Mode(rpio.Pwm)
		// pinDir.Mode(rpio.Output)

		// pinPWM.Freq(64000)
		// log.Println("PWM motor starting at speed of ", speed)
		// pinPWM.Write(rpio.High)
		// pinDir.Write(rpio.Low)
		// time.Sleep(time.Duration(duration * float64(time.Second)))
		// pinPWM.Write(rpio.Low)
		// pinDir.Write(rpio.Low)
		// log.Println("PWM motor halted")

		// log.Println("PWM motor reversing at speed of ", speed)
		// pinPWM.Write(rpio.Low)
		// pinDir.Write(rpio.High)
		// time.Sleep(time.Duration(duration * float64(time.Second)))
		// pinPWM.Write(rpio.Low)
		// pinDir.Write(rpio.Low)
		// log.Println("PWM motor halted")

		// rasp := raspi.NewAdaptor()
		// pwmPin := "26" //GPIO07
		// // dirPin := "36" //GPIO16
		// pwmDirectGPIO := gpio.NewDirectPinDriver(rasp, pwmPin)
		// log.Println("PWM motor starting at speed of ", speed)
		// pwmDirectGPIO.PwmWrite(byte(speed))
		// time.Sleep(time.Duration(duration * float64(time.Second)))
		// pwmDirectGPIO.Off()
		// log.Println("PWM motor halted")

		// pwmMotor := gpio.NewMotorDriver(rasp, pwmPin)
		// pwmMotor.DirectionPin = dirPin
		// var err error
		// work := func() {
		// 	//set speed
		// 	pwmMotor.Speed(byte(speed))
		// 	// maSpeedGpio.PwmWrite(maSpeed)
		// 	//Move forward one revolution
		// 	if err = pwmMotor.On(); err != nil {
		// 		log.Fatalln(err)
		// 	}
		// 	log.Println("PWM motor starting at speed of ", speed)
		// 	// select {
		// 	// case <-time.After(time.Duration(duration * float64(time.Second))):
		// 	time.Sleep(time.Duration(duration * float64(time.Second)))
		// 	if err = pwmMotor.Off(); err != nil {
		// 		log.Fatalln(err.Error())
		// 	}
		// 	log.Println("PWM motor halted")
		// 	// }
		// }

		// robot := gobot.NewRobot("pwmMotorBot",
		// 	[]gobot.Connection{rasp},
		// 	[]gobot.Device{pwmMotor},
		// 	work,
		// )

		// robot.Start()

		// v := make(map[int]float64)
		// v[r.pump.Pin] = speed
		// if err := r.jacks.Control(r.pump.Jack, v); err != nil {
		// 	log.Fatalln(err.Error())
		// 	return err
		// }
		// select {
		// case <-time.After(time.Duration(duration * float64(time.Second))):
		// 	v[r.pump.Pin] = 0
		// 	if err := r.jacks.Control(r.pump.Jack, v); err != nil {
		// 		log.Fatalln(err.Error())
		// 		return err
		// 	}
		// }
	}
	return nil
}

func (r *Runner) Run() {
	log.Println("doser sub system: scheduled run ", r.pump.Name)
	if err := r.Dose(r.pump.Regiment.Speed, r.pump.Regiment.Duration); err != nil {
		log.Println("ERROR: dosing sub-system. Failed to control jack. Error:", err)
		return
	}
	usage := Usage{
		Time: telemetry.TeleTime(time.Now()),
		Pump: int(r.pump.Regiment.Duration),
	}
	r.statsMgr.Update(r.pump.ID, usage)
	r.statsMgr.Save(r.pump.ID)
	//r.Telemetry().EmitMetric("doser"+r.pump.Name+"-usage", usage.Pump)
}
func (r *Runner) RunDirect(Duration float64, Speed float64) {
	log.Println("doser sub system: scheduled run ", r.pump.Name)
	if err := r.Dose(Speed, Duration); err != nil {
		log.Println("ERROR: dosing sub-system. Failed to control jack. Error:", err)
		return
	}
	usage := Usage{
		Time: telemetry.TeleTime(time.Now()),
		Pump: int(Duration),
	}
	r.statsMgr.Update(r.pump.ID, usage)
	r.statsMgr.Save(r.pump.ID)
	//r.Telemetry().EmitMetric("doser"+r.pump.Name+"-usage", usage.Pump)
}
