package system

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/reef-pi/reef-pi/controller/connectors"
	"github.com/reef-pi/reef-pi/controller/modules/ato"
	"github.com/reef-pi/reef-pi/controller/modules/doser"
	"github.com/reef-pi/reef-pi/controller/modules/equipment"
	"github.com/reef-pi/reef-pi/controller/modules/macro"
	"github.com/reef-pi/reef-pi/controller/modules/temperature"
	"github.com/reef-pi/reef-pi/controller/modules/timer"
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

	r.HandleFunc("/api/admin/program", c.GetProgram).Methods("GET")
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

func (c *Controller) GetProgram(w http.ResponseWriter, r *http.Request) {
	fn := func(string) (interface{}, error) {
		log.Println("Get program status")
		programmed, err := c.isProgrammed()
		if err != nil { //when not programmed the programmed field is not in the bucket and throws error
			return false, nil
		} else {
			return programmed, nil
		}
	}
	utils.JSONGetResponse(fn, w, r)
}

func (c *Controller) createTempSensors() (TCSIDs []string, err error) {
	TCSIDs = []string{}
	tempsubsystem, _ := c.c.Subsystem(temperature.Bucket)
	tc, ok := tempsubsystem.(*temperature.Controller)
	if !ok {
		return TCSIDs, errors.New("failed to cast temperature subsystem to temperature controller ")
	}
	sensors, err := tc.Sensors()
	if err != nil {
		return TCSIDs, err
	}
	//there must be two temp senors connected, otherwise fail
	if len(sensors) < 2 {
		return TCSIDs, errors.New("failed to detect at least two temperature sensors")
	}
	for i, sensor := range sensors {
		//create temperature controller
		tcontroller := temperature.TC{
			Name:       "temp" + strconv.Itoa(i),
			Sensor:     sensor,
			Fahrenheit: false,
			Control:    false,
			Enable:     true,
			Period:     360, //three minutes
		}
		err := tc.Create(tcontroller)
		if err != nil {
			return TCSIDs, err
		}
	}
	tcs, err := tc.List()
	for _, tc := range tcs {
		TCSIDs = append(TCSIDs, tc.ID)
	}

	return TCSIDs, nil
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
func (c *Controller) createATOs(atoNames []string,
	inlets []string,
	pumps []string) (err error, ATOIDs []string) {
	ATOIDs = []string{}
	atoSubsystem, _ := c.c.Subsystem(ato.Bucket)
	atoSub, ok := atoSubsystem.(*ato.Controller)
	if !ok {
		return errors.New("failed to cast ato subsystem to ato controller"), ATOIDs
	}
	for i, atoName := range atoNames {
		log.Println("Creating:", atoName)
		ato := ato.ATO{
			Name:   atoName,
			Inlet:  inlets[i],
			Period: 3, //3 seconds to shut off the pump
		}
		// the first two ATOs maintank and reservoir have equipments to control
		if i <= 1 {
			ato.Control = true
			ato.Pump = pumps[i]
		}
		err := atoSub.Create(ato)
		if err != nil {
			return err, ATOIDs
		}

	}
	atos, err := atoSub.List()
	if err != nil {
		return err, ATOIDs
	}
	for _, ato := range atos {
		ATOIDs = append(ATOIDs, ato.ID)
	}
	return nil, ATOIDs
}
func (c *Controller) createRemoteEquipments(outletURIs []string,
	outletNames []string, macroConfig MacroInitConfig) (equipmentIDs []string, err error) {
	equipmentIDs = []string{}
	equipmentSubsystem, _ := c.c.Subsystem(equipment.Bucket)
	equipmentSub, ok := equipmentSubsystem.(*equipment.Controller)
	if !ok {
		return equipmentIDs, errors.New("failed to cast ato subsystem to equipment controller")
	}

	for index, outletURI := range outletURIs {
		timeGuardOnCommand := ""
		if index == 0 {
			//reservoir pump outlet
			timeGuardOnCommand = fmt.Sprintf("%%3B%%20Delay%%20%d%%3B%%20Power3%%20OFF", macroConfig.mainTankATOTimeGuard*10) //the reason we multiply by 10 is tasmota takes seconds multiplied by 10, e.g., 10 seconds delay should be input as 100
		} else if index == 1 {
			//solenoid valve outlet
			timeGuardOnCommand = fmt.Sprintf("%%3B%%20Delay%%20%d%%3B%%20Power3%%20OFF", macroConfig.reservoirATOTimeGuard*10)
		}
		remoteOutlet := equipment.Equipment{
			Name:       "remote - " + outletNames[index],
			IsRemote:   true,
			OnCmd:      outletURI + "%20ON" + timeGuardOnCommand,
			OffCmd:     outletURI + "%20OFF",
			RemoteType: "http",
		}
		err := equipmentSub.Create(remoteOutlet)
		if err != nil {
			return equipmentIDs, err
		}
	}

	equipments, err := equipmentSub.List()
	if err != nil {
		return equipmentIDs, err
	}
	for _, equipment := range equipments {
		equipmentIDs = append(equipmentIDs, equipment.ID)
	}

	return equipmentIDs, nil
}
func (c *Controller) createDosingPumps(dosingPumpNames []string,
	dosingPumpPins [][]string) (err error, dosingPumpIDs []string) {
	dosingPumpIDs = []string{}

	dosingSubsystem, _ := c.c.Subsystem(doser.Bucket)
	dosingSub, ok := dosingSubsystem.(*doser.Controller)
	if !ok {
		return errors.New("failed to cast doser subsystem to doser controller"), dosingPumpIDs
	}

	for i, dosingPumpName := range dosingPumpNames {
		pump := doser.Pump{Name: dosingPumpName,
			In1Pin: dosingPumpPins[i][0],
			In2Pin: dosingPumpPins[i][1],
			In3Pin: dosingPumpPins[i][2],
			In4Pin: dosingPumpPins[i][3],
		}
		err := dosingSub.Create(pump)
		if err != nil {
			return err, dosingPumpIDs
		}
	}

	dosers, err := dosingSub.List()
	if err != nil {
		return err, dosingPumpIDs
	}
	for _, doserPump := range dosers {
		dosingPumpIDs = append(dosingPumpIDs, doserPump.ID)
	}
	return nil, dosingPumpIDs
}
func (c *Controller) marshalStruct(obj interface{}) ([]byte, error) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return bytes, err
}

func (c *Controller) createMacros(ReservoirTCSID string, ATOIDs []string, DosingIDs []string, macroConfig MacroInitConfig) (MacroIDs []string, err error) {
	MacroIDs = []string{}

	macroSubsystem, _ := c.c.Subsystem(macro.Bucket)
	macroSub, ok := macroSubsystem.(*macro.Subsystem)
	if !ok {
		return MacroIDs, errors.New("failed to cast macro subsystem to macro controller")
	}

	//begin change water macro
	//wait step -> wait for reservoir temperature to reach safe range
	waitTempStepConfig := macro.WaitTemperatureStep{
		Frequency:  macroConfig.waitTempPeriod,
		RangeTemp1: macroConfig.safeTemperatureRangeForReservoir[0],
		RangeTemp2: macroConfig.safeTemperatureRangeForReservoir[1],
		ID:         ReservoirTCSID, //reservoir temperature controller
	}
	bytes, err := c.marshalStruct(waitTempStepConfig)
	if err != nil {
		return MacroIDs, err
	}
	waitTempStep := macro.Step{
		Type:   "waittemp",
		Config: bytes,
	}

	//doser step -> pump out water from the maintank
	doserStepConfig := macro.DoserStep{
		ID:       DosingIDs[0], //doser id should be the waterchange ID
		Duration: float64(macroConfig.waterOutDuration),
		Speed:    float64(macroConfig.waterOutSpeed),
	}
	bytes, err = c.marshalStruct(doserStepConfig)
	if err != nil {
		return MacroIDs, err
	}
	doserStep := macro.Step{
		Type:   "directdoser",
		Config: bytes,
	}

	//ato step -> fill up the maintank by turning on the maintank ATO
	atoConfig := macro.GenericStep{
		ID: ATOIDs[0], //maintank ATO
		On: true,
	}
	bytes, err = c.marshalStruct(atoConfig)
	if err != nil {
		return MacroIDs, err
	}
	atoStep := macro.Step{
		Type:   "ato",
		Config: bytes,
	}

	//wait for ATO work to finish (time guard)
	waitStepConfig := macro.WaitStep{
		Duration: macroConfig.mainTankATOTimeGuard,
	}
	bytes, err = c.marshalStruct(waitStepConfig)
	if err != nil {
		return MacroIDs, err
	}
	waitStep := macro.Step{
		Type:   "wait",
		Config: bytes,
	}

	//ato step -> switch off maintank ATO
	atoConfig = macro.GenericStep{
		ID: ATOIDs[0], //maintank ATO
		On: false,
	}
	bytes, err = c.marshalStruct(atoConfig)
	if err != nil {
		return MacroIDs, err
	}
	atoOffStep := macro.Step{
		Type:   "ato",
		Config: bytes,
	}

	waterChangeSteps := []macro.Step{waitTempStep, doserStep, atoStep, waitStep, atoOffStep}
	waterChangeMacro := macro.Macro{
		Name:  "Pump Out Water and Refill Aquarium",
		Steps: waterChangeSteps,
	}
	err = macroSub.Create(waterChangeMacro)
	if err != nil {
		return MacroIDs, nil
	}
	//end water change macro

	// Begin fill up reservoir macro
	//ato step -> fill up the maintank by turning on the maintank ATO
	atoConfig = macro.GenericStep{
		ID: ATOIDs[1], //reservoir ATO
		On: true,
	}
	bytes, err = c.marshalStruct(atoConfig)
	if err != nil {
		return MacroIDs, err
	}
	atoStep = macro.Step{
		Type:   "ato",
		Config: bytes,
	}

	//wait for ATO work to finish (time guard)
	waitStepConfig = macro.WaitStep{
		Duration: macroConfig.reservoirATOTimeGuard,
	}
	bytes, err = c.marshalStruct(waitStepConfig)
	if err != nil {
		return MacroIDs, err
	}
	waitStep = macro.Step{
		Type:   "wait",
		Config: bytes,
	}

	//ato step -> switch off maintank ATO
	atoConfig = macro.GenericStep{
		ID: ATOIDs[1], //reservoir ATO
		On: false,
	}
	bytes, err = c.marshalStruct(atoConfig)
	if err != nil {
		return MacroIDs, err
	}
	atoOffStep = macro.Step{
		Type:   "ato",
		Config: bytes,
	}

	//doser step -> pump out water from the maintank
	doserStepConfig = macro.DoserStep{
		ID:       DosingIDs[1], //doser id of the conditioner pump
		Duration: float64(macroConfig.conditionerDosingDuration),
		Speed:    float64(macroConfig.conditionerDosingSpeed),
	}
	bytes, err = c.marshalStruct(doserStepConfig)
	if err != nil {
		return MacroIDs, err
	}
	doserStep = macro.Step{
		Type:   "directdoser",
		Config: bytes,
	}

	fillUpReservoirSteps := []macro.Step{atoStep, waitStep, atoOffStep, doserStep}
	fillUpReservoirMacro := macro.Macro{
		Name:  "Open Solenoid Valve to fill up reservoir (water contains chlorine), Dose 10ml Conditioner",
		Steps: fillUpReservoirSteps,
	}
	err = macroSub.Create(fillUpReservoirMacro)
	if err != nil {
		return MacroIDs, nil
	}
	// End fill up reservoir macro

	macros, err := macroSub.List()
	if err != nil {
		return MacroIDs, nil
	}
	for _, macroItem := range macros {
		MacroIDs = append(MacroIDs, macroItem.ID)
	}

	return MacroIDs, nil

}
func (c *Controller) createTimers(MacroIDs []string, timerConfig TimerInitConfig) error {
	timerSubsystem, _ := c.c.Subsystem(timer.Bucket)
	timerController, ok := timerSubsystem.(*timer.Controller)
	if !ok {
		return errors.New("failed to cast timer subsystem to timer controller")
	}
	for _, schedule := range timerConfig.waterChangeSchedules {
		err := timerController.Create(schedule)
		if err != nil {
			return err
		}
	}
	err := timerController.Create(timerConfig.reservoirFillUpSchedule)
	if err != nil {
		return err
	}
	return nil
}
func (c *Controller) Program(w http.ResponseWriter, r *http.Request) {
	fn := func(string) (interface{}, error) {
		log.Println("Programming reef-pi controller")
		//check if programmed already and return if so
		if programmed, err := c.isProgrammed(); err == nil && programmed {
			return "Already programmed", nil
		}
		//Create temperature controllers (at least two sensors must be detected)
		TCSIDs, err := c.createTempSensors()
		if err != nil {
			log.Fatalln(err)
		}
		//Create water level sensors (at least three)
		inletPins := []string{"10", "9", "27", "22"} //BCM pins
		err, inletIDs := c.createWaterLevelSensors(inletPins)
		if err != nil {
			log.Fatalln(err)
		}

		//TODO: Hardcoded init values, need to be configured based on fish types, reservoir size, and pump calibration times
		macroConfig := MacroInitConfig{
			safeTemperatureRangeForReservoir: []float64{28.5, 30.0},
			waitTempPeriod:                   360,
			waterOutDuration:                 140,
			waterOutSpeed:                    100,
			mainTankATOTimeGuard:             25,
			reservoirATOTimeGuard:            420,
			conditionerDosingDuration:        5,
			conditionerDosingSpeed:           70,
		}

		//Create equipments (maintank heater, reservoir heater, reservoir pump, solenoid valve)
		outletURIs := []string{
			"http://192.168.178.57/cm?cmnd=Backlog%20Power2", //reservoir pump
			"http://192.168.178.57/cm?cmnd=Backlog%20Power3", //solenoid valve
			"http://192.168.178.57/cm?cmnd=Backlog%20Power4", //maintank heater
			"http://192.168.178.57/cm?cmnd=Backlog%20Power5", //reservoir heater
		}
		outletNames := []string{
			"reservoir pump",
			"solenoid valve",
			"maintank heater",
			"reservoir heater",
		}
		equipmentIDs, err := c.createRemoteEquipments(outletURIs, outletNames, macroConfig)
		if err != nil {
			log.Fatalln(err)
		}
		// Create ATOs (maintank, reservoir, and conditioner jar)
		// The conditioner ATO doesn't control any outlet, but issues warning when the jar gets empty
		atoNames := []string{"maintank", "reservoir", "conditioner"} //must be in the same order as inletPins
		// We need solenoid valve ID and reservoir pump ID
		atoEquipmentIDs := equipmentIDs[:2]
		if len(atoEquipmentIDs) != 2 {
			log.Fatalln(errors.New("Didn't get two equipment IDs for ATOs"))
		}
		err, ATOIDs := c.createATOs(atoNames, inletIDs, atoEquipmentIDs)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(ATOIDs)

		//Create dosing pumps
		dosingPumpNames := []string{"waterchange-pump", "conditioner-pump"}
		dosingPumpPins := [][]string{{"16", "", "7", "1000"}, {"25", "24", "23", "1000"}} //hardcoded based on the PCB
		err, DosingIDs := c.createDosingPumps(dosingPumpNames, dosingPumpPins)
		log.Println(DosingIDs)

		//Create macros -> needs ato ids and dosing pump ids
		MacroIDs, err := c.createMacros(TCSIDs[1] /*reservoir tcs*/, ATOIDs, DosingIDs, macroConfig)
		log.Println(MacroIDs)

		//Create timers -> needs macroIds

		//water change timer configs
		targetConfig := timer.TriggerMacro{
			ID: MacroIDs[0], //water change macro
		}
		bytes, err := c.marshalStruct(targetConfig)
		if err != nil {
			log.Fatalln(err)
		}
		waterChangeSchedule := timer.Job{
			Name:   "waterchange 9AM",
			Type:   "macro",
			Enable: true,
			Month:  "*",
			Week:   "*",
			Day:    "*",
			Hour:   "9",
			Minute: "0",
			Second: "0",
			Target: bytes,
		}

		// reservoir fillup timer configs
		targetConfig = timer.TriggerMacro{
			ID: MacroIDs[1], //reservoir fill up macro
		}
		bytes, err = c.marshalStruct(targetConfig)
		if err != nil {
			log.Fatalln(err)
		}
		fillUpReservoirSchedule := timer.Job{
			Name:   "Fill up Reservoir every 3 days @9PM",
			Type:   "macro",
			Enable: true,
			Month:  "*",
			Week:   "*",
			Day:    "*/3",
			Hour:   "21",
			Minute: "0",
			Second: "0",
			Target: bytes,
		}

		timerConfig := TimerInitConfig{
			waterChangeSchedules:    []timer.Job{waterChangeSchedule},
			reservoirFillUpSchedule: fillUpReservoirSchedule,
		}

		err = c.createTimers(MacroIDs, timerConfig)

		if err != nil {
			return "", fmt.Errorf("Failed to program. Error: " + err.Error())
		}

		err = c.markAsProgrammed()
		if err != nil {
			return "", fmt.Errorf("Failed to program. Error: " + err.Error())
		}

		return "GUUT", nil
	}
	utils.JSONGetResponse(fn, w, r)
}
