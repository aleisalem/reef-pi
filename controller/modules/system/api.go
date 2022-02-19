package system

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/reef-pi/reef-pi/controller/connectors"
	"github.com/reef-pi/reef-pi/controller/modules/temperature"
	"github.com/reef-pi/reef-pi/controller/utils"
)

func (c *Controller) LoadAPI(r *mux.Router) {
	r.HandleFunc("/api/display/on", c.EnableDisplay).Methods("POST")
	r.HandleFunc("/api/display/off", c.DisableDisplay).Methods("POST")
	r.HandleFunc("/api/display", c.SetBrightness).Methods("POST")
	r.HandleFunc("/api/display", c.GetDisplayState).Methods("GET")
	r.HandleFunc("/api/admin/poweroff", c.Poweroff).Methods("POST")
	r.HandleFunc("/api/admin/reboot", c.Reboot).Methods("POST")
	r.HandleFunc("/api/admin/reload", c.reload).Methods("POST")
	r.HandleFunc("/api/admin/upgrade", c.upgrade).Methods("POST")
	r.HandleFunc("/api/info", c.GetSummary).Methods("GET")

	r.HandleFunc("/api/admin/program", c.Program).Methods("POST")

	if c.config.Pprof {
		c.enablePprof(r)
	}
}

func (c *Controller) enablePprof(r *mux.Router) {
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func (c *Controller) EnableDisplay(w http.ResponseWriter, r *http.Request) {
	fn := func(_ string) error {
		return c.enableDisplay()
	}
	utils.JSONDeleteResponse(fn, w, r)
}

func (c *Controller) DisableDisplay(w http.ResponseWriter, r *http.Request) {
	fn := func(_ string) error {
		return c.disableDisplay()
	}
	utils.JSONDeleteResponse(fn, w, r)
}

func (c *Controller) SetBrightness(w http.ResponseWriter, r *http.Request) {
	var conf DisplayConfig
	fn := func() error {
		return c.setBrightness(conf.Brightness)
	}
	utils.JSONCreateResponse(&conf, fn, w, r)
}

func (c *Controller) GetDisplayState(w http.ResponseWriter, r *http.Request) {
	fn := func(id string) (interface{}, error) {
		if !c.config.Display {
			return DisplayState{}, nil
		}
		return c.currentDisplayState()
	}
	utils.JSONGetResponse(fn, w, r)
}
func (t *Controller) GetSummary(w http.ResponseWriter, r *http.Request) {
	fn := func(id string) (interface{}, error) {
		return t.ComputeSummary(), nil
	}
	utils.JSONGetResponse(fn, w, r)
}

func (c *Controller) Poweroff(w http.ResponseWriter, r *http.Request) {
	fn := func(string) (interface{}, error) {
		log.Println("Shutting down reef-pi controller")
		out, err := utils.Command("/bin/systemctl", "poweroff").WithDevMode(c.config.DevMode).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("Failed to power off reef-pi. Output:" + string(out) + ". Error: " + err.Error())
		}
		return out, nil
	}
	utils.JSONGetResponse(fn, w, r)
}

func (c *Controller) Reboot(w http.ResponseWriter, r *http.Request) {
	fn := func(string) (interface{}, error) {
		log.Println("Rebooting reef-pi controller")
		out, err := utils.Command("/bin/systemctl", "reboot").WithDevMode(c.config.DevMode).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("Failed to reboot reef-pi. Output:" + string(out) + ". Error: " + err.Error())
		}
		return out, nil
	}
	utils.JSONGetResponse(fn, w, r)
}

func (c *Controller) reload(w http.ResponseWriter, r *http.Request) {
	fn := func(string) (interface{}, error) {
		log.Println("Reloading reef-pi controller")
		out, err := utils.Command("/bin/systemctl", "restart", "reef-pi.service").WithDevMode(c.config.DevMode).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("Failed to reload reef-pi. Output:" + string(out) + ". Error: " + err.Error())
		}
		return out, nil
	}
	utils.JSONGetResponse(fn, w, r)
}
func (c *Controller) upgrade(w http.ResponseWriter, r *http.Request) {
	fn := func(string) (interface{}, error) {
		log.Println("Upgrading reef-pi controller")
		err := utils.SystemdExecute("/usr/bin/apt-get update -y")
		if err != nil {
			return "", fmt.Errorf("Failed to update. Error: " + err.Error())
		}
		return "", nil
	}
	utils.JSONGetResponse(fn, w, r)
}

func (c *Controller) createTempSensors() error {
	tempsubsystem, _ := c.c.Subsystem(temperature.Bucket)
	tc, ok := tempsubsystem.(*temperature.Controller)
	if !ok {
		return errors.New("failed to cast temperature subsystem to temperature controller ")
	}
	sensors, err := tc.Sensors()
	if err != nil {
		return err
	}
	//there must be two temp senors connected, otherwise fail
	if len(sensors) < 2 {
		return errors.New("failed to detect at least two temperature sensors")
	}
	for i, sensor := range sensors {
		//create temperature controller
		tcontroller := temperature.TC{
			Name:       "temp" + string(i),
			Sensor:     sensor,
			Fahrenheit: false,
			Control:    false,
			Enable:     true}
		err := tc.Create(tcontroller)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *Controller) createWaterLevelSensors(inletPins []string) (err error, inletIDs []string) {
	inletIDs = []string{}
	//Create input jacks
	for _, pin := range inletPins {
		intPin, err := strconv.Atoi(pin)
		if err != nil {
			return errors.New("failed to convert input pin to integer " + err.Error()), inletIDs
		}
		inlet := connectors.Inlet{Name: "inletPin" + string(pin), Pin: intPin, Driver: "rpi"}
		err = c.c.DM().Inlets().Create(inlet)
		if err != nil {
			return errors.New("failed to create inlet " + err.Error()), inletIDs
		}
	}
	inlets, err := c.c.DM().Inlets().List()
	for _, inlet := range inlets {
		inletIDs = append(inletIDs, inlet.ID)
	}

	return nil, inletIDs
}
func (c *Controller) createATOs(atoNames []string, inlets []string, pumps []string) error {
	// atoSubsystem, _ := c.c.Subsystem(ato.Bucket)
	// atoSub, ok := atoSubsystem.(*ato.Controller)
	// if !ok {
	// 	return errors.New("failed to cast ato subsystem to ato controller")
	// }
	// for i, atoName := range atoNames {
	// 	ato := ato.ATO{Name: atoName, Inlet: inlets[i]}
	// 	atoSub.Create()
	// }
	return nil
}
func (c *Controller) Program(w http.ResponseWriter, r *http.Request) {
	fn := func(string) (interface{}, error) {
		log.Println("Programming reef-pi controller")
		//check if programmed already and return if so

		//if not programmed
		//Create temperature controllers (at least two)
		err := c.createTempSensors()
		if err != nil {
			log.Fatalln(err)
		}
		// //Create water level sensors (at least three)
		// inletPins := []string{"10", "9", "27", "22"} //BCM pins
		// err, inletIDs := c.createWaterLevelSensors(inletPins)
		// if err != nil {
		// 	log.Fatalln(err)
		// }

		// //Create equipments (maintank heater, reservoir heater, reservoir pump, solenoid valve)
		// outletPins := []string{"", "", "", ""}

		// //Create ATOs (maintank, reservoir, and conditioner jar)
		// atoNames := []string{"maintank", "reservoir", "conditioner"} //must be in the same order as inletPins
		// err = c.createATOs(atoNames, inletIDs, pumpIDs)
		// //Create doing pumps

		// //Create macros

		// //Create timers

		if err != nil {
			return "", fmt.Errorf("Failed to program. Error: " + err.Error())
		}
		return "", nil
	}
	utils.JSONGetResponse(fn, w, r)
}
