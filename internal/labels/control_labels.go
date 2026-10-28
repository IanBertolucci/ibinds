package labels

// controlLabels maps iRacing's internal control ids to the names shown in the
// in-game controls screen. The source list is binds_do_iracing.txt; ids absent
// from that list intentionally keep their technical name.
//
// The following entries are inferred rather than directly confirmed: LFE vs
// TrueForce (assigned TrueForce to Wheel LFE and LFE to BassShaker), the
// WeightJacker -> spring set mapping, RpyIncreaseRate -> Slow Motion, and
// DCamVanishY -> Shift Horizon. All other mappings follow the in-game list.
var controlLabels = buildControlLabels()

func buildControlLabels() map[string]string {
	m := map[string]string{
		// Cockpit
		"Reset":                   "Enter / Exit / Tow Car",
		"LookLeft":                "Look Left",
		"LookRight":               "Look Right",
		"LookUp":                  "Look Up",
		"LookDown":                "Look Down",
		"Handbrake":               "Handbrake",
		"Throttle2":               "Second Throttle",
		"Brake2":                  "Second Brake",
		"Clutch2":                 "Second Clutch",
		"ShiftUp2":                "Second Upshift",
		"ShiftDown2":              "Second Downshift",
		"Ignition":                "Ignition",
		"Starter":                 "Starter",
		"HeadlightFlash":          "Headlight Flash",
		"PitSpeedLimiter":         "Pit Speed Limiter",
		"PushToPass":              "Push To Pass / Overtake",
		"DRS":                     "DRS",
		"TractionControlArm":      "Traction Control Arm",
		"FuelToEndToggle":         "Autofuel Toggle",
		"FuelToEndMarginInc":      "Autofuel Lap Margin Inc",
		"FuelToEndMarginDec":      "Autofuel Lap Margin Dec",
		"TearOffVisor":            "Tear Off Visor",
		"ToggleWindshieldWipers":  "Toggle Windshield Wipers",
		"TriggerWindshieldWipers": "Trigger Windshield Wipers",
		"DashBoxToggle":           "Toggle Dash Box",
		"PitSummaryToggle":        "Toggle On-Track Pit Summary",

		// Driver's view
		"DCamVanishYInc":             "Shift Horizon Up",
		"DCamVanishYDec":             "Shift Horizon Down",
		"DCamDrvHeightInc":           "Increase Driver Height",
		"DCamDrvHeightDec":           "Decrease Driver Height",
		"DCamFOVInc":                 "Increase Driver Field of View",
		"DCamFOVDec":                 "Decrease Driver Field of View",
		"RecenterHeadMountedDisplay": "Recenter VR View",

		// Interface
		"ToggleSpeedGearDisplay":   "Speed / Gear / Pedals Display",
		"ToggleRadioDisplay":       "Radio Display",
		"ToggleSystemMeterDisplay": "FPS / Network Display",
		"ToggleVideoTimerDisplay":  "Video Timer",
		"ReportLatency":            "Report Latency",
		"ToggleRadar":              "Toggle Weather Radar",
		"ToggleVirtualMirror":      "Toggle Virtual Mirror",
		"ToggleDrivingLine":        "Toggle Driving Line",
		"ToggleEditLayoutMode":     "Toggle UI Edit",
		"ScaleUIUp":                "Scale UI Up",
		"ScaleUIDown":              "Scale UI Down",
		"ToggleUIVisible":          "Cycle UI Visibility Mode",

		// Black box
		"BlackBoxShowNext": "Next Black Box",
		"BlackBoxShowPrev": "Previous Black Box",
		"BlackBoxF1":       "Lap Timing Black Box",
		"BlackBoxF2":       "Standings Black Box",
		"BlackBoxF3":       "Relative Black Box",
		"BlackBoxF4":       "Fuel Black Box",
		"BlackBoxF5":       "Tires Black Box",
		"BlackBoxF6":       "Tire Info Black Box",
		"BlackBoxF7":       "Pit-stop Adjustments Black Box",
		"BlackBoxF8":       "In-car Adjustments Black Box",
		"BlackBoxF9":       "Mirror Adjustments Black Box",
		"BlackBoxF10":      "Radio Adjustments Black Box",
		"BlackBoxF11":      "Weather Black Box",
		"BlackBoxNext":     "Select Next Control",
		"BlackBoxPrev":     "Select Previous Control",
		"BlackBoxInc":      "Increment Selected Control",
		"BlackBoxDec":      "Decrement Selected Control",
		"BlackBoxToggle":   "Toggle Selected Control",
		"BlackBoxSize":     "Toggle Black Box Size",

		// Split-time delta
		"SplitsDeltaNext": "Next Splits Delta Display",
		"SplitsDeltaPrev": "Prev Splits Delta Display",
		"ToggleRefCar":    "Toggle Display Reference Car",

		// Spotter
		"SPCCVolumeUp":          "Spotter Louder",
		"SPCCVolumeDown":        "Spotter Quieter",
		"SPCCShutUp":            "Spotter Silence",
		"SPCCDamageReport":      "Spotter Damage Report",
		"SPCCWeatherReport":     "Spotter Weather Report",
		"SPCCToggleLapTimes":    "Toggle Report Laps",
		"SPCCAnounceLeaderLine": "Announce Leader at S/F Line",

		// Advanced analysis
		"CustSectSetStart":       "Mark Start Point",
		"CustSectSetEnd":         "Mark End Point",
		"SaveActiveReset":        "Save Start Point",
		"RunActiveReset":         "Reset to Start Point",
		"ToggleirsdkDiskLogging": "Toggle Telemetry Logging",
		"IrsdkDriverMarker":      "Mark Event in Telemetry",

		// Audio & force feedback
		"MasterVolumeUp":      "Master Volume Louder",
		"MasterVolumeDown":    "Master Volume Quieter",
		"AutoFFB":             "Auto Compute FFB Force",
		"IncFFB":              "Increase FFB Force",
		"DecFFB":              "Decrease FFB Force",
		"LFEVolumeUp":         "BassShaker LFE Louder",
		"LFEVolumeDown":       "BassShaker LFE Quieter",
		"TrueForceVolumeUp":   "Wheel LFE Louder",
		"TrueForceVolumeDown": "Wheel LFE Quieter",

		// Chat
		"VChatPushToTalk":  "Push-to-Talk",
		"VChatChannelNext": "Next Channel",
		"VChatChannelPrev": "Previous Channel",
		"VChatVolumeUp":    "Voice Chat Louder",
		"VChatVolumeDown":  "Voice Chat Quieter",
		"VChatMute":        "Mute Voice Chat",
		"VChatDriverMute":  "Mute a Driver",
		"TChatInitiate":    "Text Chat",
		"TChatReply":       "Text Chat Reply",
		"TChatToggle":      "Text Chat Toggle",
		"AutoChat1":        "Pitting In",
		"AutoChat2":        "Pitting Out",
		"AutoChat3":        "Pass Left",
		"AutoChat4":        "Pass Right",
		"AutoChat5":        "Sorry!",
		"AutoChat6":        "Thanks!",
		"AutoChat8":        ":(",
		"AutoChat12":       "#rf rr",
		"AutoChat13":       "#lf rf lr rr",
		"AutoChat14":       "#fuel +0.5l",
		"AutoChat15":       "#fuel -0.5l",

		// Camera & screen capture
		"TriggerScreenshotCapture":      "Screenshot Capture",
		"TriggerGiantScreenshotCapture": "Giant Screenshot Capture",
		"ToggleVideoCapture":            "Video Capture",
		"CamToggleCamTool":              "Open Camera Tool",

		// Replay
		"RpyPausePlay":    "Play / Pause",
		"RpyStop":         "Stop",
		"RpyFF":           "Fast Forward",
		"RpyRew":          "Rewind",
		"RpyIncreaseRate": "Slow Motion",
		"RpyFrameAdv":     "Next Frame",
		"RpyFrameRew":     "Previous Frame",
		"RpySessionFF":    "Next Session",
		"RpySessionRew":   "Previous Session",
		"RpyLapFF":        "Next Lap",
		"RpyLapRew":       "Previous Lap",
		"RpyIncidentFF":   "Next Incident",
		"RpyIncidentRew":  "Previous Incident",
		"RpyToStart":      "Start of Recording",
		"RpyToEnd":        "End of Recording",
		"NextRpyCamCam":   "Next Camera",
		"PrevRpyCamCam":   "Previous Camera",
		"NextRpyCamGroup": "Next Sub Camera",
		"PrevRpyCamGroup": "Previous Sub Camera",
		"FocusNextCar":    "Next Car",
		"FocusPrevCar":    "Previous Car",
		"FocusPlayerCar":  "Your Car",
		"NextDrivingCam":  "Next Driving Camera",
		"PrevDrivingCam":  "Previous Driving Camera",

		// Car setup toggles
		"ABSToggle":             "ABS Toggle",
		"RFBrakeAttachedToggle": "RF Brake Attached Toggle",
		"FCYToggle":             "FCY Mode Toggle",
		"FuelNoCutToggle":       "Disable Fuel Cut",
		"LowFuelAccept":         "Low Fuel Accept",
		"TractionControlToggle": "Traction Control Toggle",
		"InLapToggle":           "In Lap Mode Toggle",
		"HysBoostHold":          "HYS Boost Hold",
		"HysRegenHold":          "HYS Regen Hold",
		"HysNoBoostToggle":      "HYS No Boost Toggle",
		"ReloadCarTexture":      "Reload Car Texture",
	}

	add := func(base, inc, dec, level string) {
		m[inc] = base + " ↑"
		m[dec] = base + " ↓"
		if level != "" {
			m[level] = base
		}
	}

	// Brakes
	add("ABS Set", "ABSInc", "ABSDec", "ABSLevel")
	add("Brake Bias Set", "BrakeBiasInc", "BrakeBiasDec", "BrakeBiasLevel")
	add("Brake Bias Fine Set", "BrakeBiasFineInc", "BrakeBiasFineDec", "BrakeBiasFineLevel")
	add("Peak Brake Bias Set", "PeakBrakeBiasInc", "PeakBrakeBiasDec", "PeakBrakeBiasLevel")
	add("Brake Misc. Set", "BrakeMiscInc", "BrakeMiscDec", "BrakeMiscLevel")

	// Engine & throttle
	add("Engine Braking Set", "EngineBrakingInc", "EngineBrakingDec", "EngineBrakingLevel")
	add("Engine Power Set", "EnginePowerInc", "EnginePowerDec", "EnginePowerLevel")
	add("Throttle Shaping Set", "ThrottleShapeInc", "ThrottleShapeDec", "ThrottleShapeLevel")
	add("Boost Level Set", "BoostLevelInc", "BoostLevelDec", "BoostLevelLevel")
	add("Launch RPM Set", "LaunchRPMInc", "LaunchRPMDec", "LaunchRPMLevel")

	// Fuel
	add("Fuel Mixture Set", "FuelMixtureInc", "FuelMixtureDec", "FuelMixtureLevel")
	add("Fuel Cut Position Set", "FuelCutPositionInc", "FuelCutPositionDec", "FuelCutPositionLevel")

	// Shocks & springs
	add("Left Spring Set", "WeightJackerLeftInc", "WeightJackerLeftDec", "WeightJackerLeftLevel")
	add("Right Spring Set", "WeightJackerRightInc", "WeightJackerRightDec", "WeightJackerRightLevel")
	add("Left Front Shock Set", "ShockLeftFrontInc", "ShockLeftFrontDec", "ShockLeftFrontLevel")
	add("Right Front Shock Set", "ShockRightFrontInc", "ShockRightFrontDec", "ShockRightFrontLevel")
	add("Left Rear Shock Set", "ShockLeftRearInc", "ShockLeftRearDec", "ShockLeftRearLevel")
	add("Right Rear Shock Set", "ShockRightRearInc", "ShockRightRearDec", "ShockRightRearLevel")

	// Traction control
	add("Traction Control Set", "TractionControlInc", "TractionControlDec", "TractionControlLevel")
	add("Traction Control 2 Set", "TractionControl2Inc", "TractionControl2Dec", "TractionControl2Level")
	add("Traction Control 3 Set", "TractionControl3Inc", "TractionControl3Dec", "TractionControl3Level")
	add("Traction Control 4 Set", "TractionControl4Inc", "TractionControl4Dec", "TractionControl4Level")

	// Differential
	add("Differential Preload Set", "DiffPreloadInc", "DiffPreloadDec", "DiffPreloadLevel")
	add("Differential Entry Set", "DiffEntryInc", "DiffEntryDec", "DiffEntryLevel")
	add("Differential Middle Set", "DiffMiddleInc", "DiffMiddleDec", "DiffMiddleLevel")
	add("Differential Exit Set", "DiffExitInc", "DiffExitDec", "DiffExitLevel")

	// Anti-roll bars & steering
	add("Front Anti-Roll Bar Set", "AntiRollFrontInc", "AntiRollFrontDec", "AntiRollFrontLevel")
	add("Rear Anti-Roll Bar Set", "AntiRollRearInc", "AntiRollRearDec", "AntiRollRearLevel")
	add("Power Steering Set", "PowerSteeringInc", "PowerSteeringDec", "PowerSteeringLevel")

	// Dash
	add("Dash Page Set", "DashPageInc", "DashPageDec", "DashPageLevel")
	add("Dash Page 2 Set", "DashPage2Inc", "DashPage2Dec", "DashPage2Level")

	// Aerodynamics
	add("Front Wing Set", "WingFrontInc", "WingFrontDec", "WingFrontLevel")
	add("Rear Wing Set", "WingRearInc", "WingRearDec", "WingRearLevel")
	add("Qualifying Tape Set", "QTapeInc", "QTapeDec", "QTapeLevel")

	// Hybrid
	add("MGU-K Re-Gen Gain Set", "MGUKRegenGainInc", "MGUKRegenGainDec", "MGUKRegenGainLevel")
	add("MGU-K Deploy Mode Set", "MGUKDeployModeInc", "MGUKDeployModeDec", "MGUKDeployModeLevel")
	add("MGU-K Fixed Deploy Set", "MGUKDeployFixedInc", "MGUKDeployFixedDec", "MGUKDeployFixedLevel")

	return m
}
