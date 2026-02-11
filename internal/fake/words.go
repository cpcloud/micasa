// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package fake

// ---------------------------------------------------------------------------
// Project titles -- per project type
// ---------------------------------------------------------------------------

var projectTitles = map[string][]string{
	"Appliance": {
		"Replace garbage disposal",
		"Install range hood",
		"Upgrade dishwasher",
		"Replace microwave",
		"Install water softener",
		"Replace oven element",
	},
	"Electrical": {
		"Upgrade electrical panel",
		"Install recessed lighting",
		"Add GFCI outlets in bathroom",
		"Replace ceiling fan",
		"Install EV charger outlet",
		"Run ethernet to home office",
		"Add outdoor floodlights",
	},
	"Exterior": {
		"Replace back deck boards",
		"Paint exterior trim",
		"Repair fence gate",
		"Build raised garden beds",
		"Install patio pavers",
		"Seal concrete driveway",
		"Replace mailbox post",
	},
	"Flooring": {
		"Refinish hardwood floors",
		"Install LVP in basement",
		"Replace bathroom tile",
		"Add carpet runner on stairs",
		"Tile entryway",
		"Polish concrete garage floor",
	},
	"HVAC": {
		"Replace furnace blower motor",
		"Install programmable thermostat",
		"Add attic insulation",
		"Seal ductwork joints",
		"Install mini-split in sunroom",
		"Clean and inspect chimney",
	},
	"Landscaping": {
		"Front yard landscaping",
		"Install drip irrigation",
		"Remove dead oak tree",
		"Plant privacy hedge",
		"Build retaining wall",
		"Overseed and aerate lawn",
		"Mulch all beds",
	},
	"Painting": {
		"Paint master bedroom",
		"Paint kitchen cabinets",
		"Paint front door",
		"Touch up hallway scuffs",
		"Stain deck",
		"Paint basement ceiling",
	},
	"Plumbing": {
		"Replace water heater",
		"Fix leaky kitchen faucet",
		"Install bidet attachment",
		"Replace shower valve",
		"Repipe galvanized supply lines",
		"Snake main sewer line",
		"Replace hose bibbs",
	},
	"Remodel": {
		"Kitchen countertop upgrade",
		"Finish basement",
		"Convert closet to pantry",
		"Add mudroom bench",
		"Bathroom vanity replacement",
		"Open up kitchen wall",
	},
	"Roof": {
		"Replace missing shingles",
		"Install gutter guards",
		"Repair flashing around chimney",
		"Add ridge vent",
		"Full roof replacement",
		"Patch flat roof section",
	},
	"Structural": {
		"Repair cracked foundation wall",
		"Sister floor joists in basement",
		"Replace rotted rim joist",
		"Add support post in crawlspace",
		"Fix sagging beam",
	},
	"Windows": {
		"Replace front windows",
		"Install storm windows",
		"Fix broken window crank",
		"Add window film for UV",
		"Replace sliding glass door",
		"Install skylight",
	},
}

// ---------------------------------------------------------------------------
// Appliances
// ---------------------------------------------------------------------------

var applianceNames = []string{
	"Refrigerator",
	"Washer",
	"Dryer",
	"Dishwasher",
	"Water Heater",
	"Tankless Water Heater",
	"Furnace",
	"Central AC",
	"Mini-Split AC",
	"Oven / Range",
	"Microwave",
	"Garage Door Opener",
	"Sump Pump",
	"Water Softener",
	"Garbage Disposal",
	"Dehumidifier",
	"Whole-House Fan",
	"Smoke / CO Detector",
	"Thermostat",
	"Ceiling Fan",
}

var applianceBrands = []string{
	"Frostline",
	"CleanWave",
	"AquaMax",
	"AirComfort",
	"LiftRight",
	"BrightHome",
	"CoolBreeze",
	"SteadyHeat",
	"QuietFlow",
	"PureAir",
	"IronGuard",
	"ClearView",
}

var applianceLocations = []string{
	"Kitchen",
	"Laundry Room",
	"Basement",
	"Garage",
	"Utility Closet",
	"Bathroom",
	"Master Bedroom",
	"Living Room",
	"Attic",
	"Hallway",
	"Sunroom",
	"Crawlspace",
}

// ---------------------------------------------------------------------------
// Vendors
// ---------------------------------------------------------------------------

var vendorTrades = []string{
	"Plumbing",
	"Electric",
	"Landscaping",
	"Roofing",
	"HVAC",
	"Painting",
	"Handyman",
	"Flooring",
	"Fencing",
	"Pest Control",
	"Window",
	"Concrete",
}

var vendorSuffixes = []string{
	"Services",
	"Solutions",
	"Co",
	"Pros",
	"Works",
	"Group",
}

var vendorAdjectives = []string{
	"Premier",
	"Central",
	"Reliable",
	"Bright",
	"Quality",
	"Summit",
	"Eagle",
	"Heritage",
	"Greenleaf",
	"Sparks",
	"Hartley",
	"Apex",
}

// ---------------------------------------------------------------------------
// Maintenance items -- per category
// ---------------------------------------------------------------------------

var maintenanceItems = map[string][]struct {
	Name     string
	Interval int // months
	Notes    string
}{
	"Appliance": {
		{Name: "Refrigerator coil cleaning", Interval: 6, Notes: "Vacuum coils underneath"},
		{Name: "Dishwasher filter cleaning", Interval: 1, Notes: "Rinse under running water"},
		{Name: "Dryer vent cleaning", Interval: 12, Notes: "Full duct run to exterior"},
		{Name: "Garbage disposal cleaning", Interval: 2, Notes: "Ice cubes and citrus peels"},
		{Name: "Range hood filter soak", Interval: 3, Notes: "Degrease in hot soapy water"},
	},
	"Electrical": {
		{Name: "Test GFCI outlets", Interval: 6, Notes: "Press test/reset on each"},
		{
			Name:     "Inspect panel for corrosion",
			Interval: 12,
			Notes:    "Visual check, look for scorch marks",
		},
		{Name: "Check outdoor lighting", Interval: 6, Notes: "Replace bulbs, clean fixtures"},
	},
	"Exterior": {
		{Name: "Gutter cleaning", Interval: 6, Notes: "Front and back, check downspout screens"},
		{Name: "Power wash siding", Interval: 12, Notes: "Low pressure, top down"},
		{
			Name:     "Inspect caulking around windows",
			Interval: 12,
			Notes:    "Re-caulk any cracked spots",
		},
		{Name: "Clean and seal deck", Interval: 12, Notes: "Sand rough spots first"},
		{Name: "Check weather stripping", Interval: 12, Notes: "All exterior doors"},
	},
	"HVAC": {
		{Name: "HVAC filter replacement", Interval: 3, Notes: "MERV 13, buy in bulk"},
		{Name: "Furnace annual inspection", Interval: 12, Notes: "Schedule in late summer"},
		{Name: "AC condenser cleaning", Interval: 12, Notes: "Hose down coils, clear debris"},
		{Name: "Check thermostat batteries", Interval: 6, Notes: "Replace if low"},
		{Name: "Bleed radiators", Interval: 12, Notes: "Start at top floor"},
	},
	"Interior": {
		{Name: "Touch up interior paint", Interval: 12, Notes: "Keep extra cans in garage"},
		{Name: "Deep clean carpets", Interval: 12, Notes: "Rent extractor or hire out"},
		{Name: "Clean bathroom exhaust fans", Interval: 6, Notes: "Remove cover, vacuum motor"},
	},
	"Landscaping": {
		{Name: "Lawn mower blade sharpening", Interval: 12, Notes: "Or replace if nicked"},
		{Name: "Aerate lawn", Interval: 12, Notes: "Best in fall for cool-season grass"},
		{Name: "Prune trees and shrubs", Interval: 12, Notes: "Late winter for most species"},
		{Name: "Winterize irrigation", Interval: 12, Notes: "Blow out lines before first freeze"},
		{Name: "Fertilize lawn", Interval: 3, Notes: "Follow local extension recs"},
	},
	"Plumbing": {
		{Name: "Water softener salt refill", Interval: 2, Notes: "40lb bag solar salt"},
		{Name: "Sump pump test", Interval: 6, Notes: "Pour 5 gallons, confirm auto-start"},
		{Name: "Flush water heater", Interval: 12, Notes: "Drain sediment from tank bottom"},
		{Name: "Check supply line hoses", Interval: 12, Notes: "Replace any bulging braided lines"},
		{Name: "Test water pressure", Interval: 12, Notes: "Should be 40-60 PSI at hose bib"},
	},
	"Safety": {
		{Name: "Smoke detector batteries", Interval: 12, Notes: "Use 9V lithium"},
		{Name: "CO detector test", Interval: 6, Notes: "Press test button on each unit"},
		{Name: "Fire extinguisher check", Interval: 12, Notes: "Verify gauge in green zone"},
		{Name: "Radon test", Interval: 24, Notes: "Use charcoal canister kit"},
	},
	"Structural": {
		{Name: "Inspect foundation cracks", Interval: 12, Notes: "Mark and measure any changes"},
		{Name: "Check attic for leaks", Interval: 6, Notes: "Look after heavy rain"},
		{Name: "Inspect crawlspace moisture", Interval: 12, Notes: "Check vapor barrier condition"},
	},
}

// ---------------------------------------------------------------------------
// House profile
// ---------------------------------------------------------------------------

var foundationTypes = []string{
	"Poured Concrete", "Block", "Crawlspace", "Slab", "Pier and Beam", "Stone",
}

var wiringTypes = []string{
	"Copper", "Aluminum", "Knob and Tube", "Romex NM-B",
}

var roofTypes = []string{
	"Asphalt Shingle", "Metal Standing Seam", "Clay Tile", "Slate", "Wood Shake", "TPO Flat",
}

var exteriorTypes = []string{
	"Vinyl Siding", "Brick", "Stucco", "Wood Clapboard", "Fiber Cement", "Stone Veneer",
}

var heatingTypes = []string{
	"Forced Air Gas", "Heat Pump", "Radiant Floor", "Boiler / Radiator", "Electric Baseboard",
}

var coolingTypes = []string{
	"Central AC", "Mini-Split", "Window Units", "Evaporative Cooler", "Heat Pump",
}

var waterSources = []string{
	"Municipal", "Well", "Community Well",
}

var sewerTypes = []string{
	"Municipal", "Septic", "Cesspool",
}

var parkingTypes = []string{
	"Attached 2-Car", "Detached 1-Car", "Carport", "Street Only", "Attached 1-Car", "Detached 2-Car",
}

var basementTypes = []string{
	"Finished", "Unfinished", "Partial", "Crawlspace", "None",
}

var insuranceCarriers = []string{
	"Acme Insurance", "Heritage Mutual", "Lakewood Insurance", "Redwood Underwriters",
	"Summit Coverage", "Bayfield Insurance", "Elm Street Insurance", "Crestview Insurance",
}

// ---------------------------------------------------------------------------
// Service log
// ---------------------------------------------------------------------------

var serviceLogNotes = []string{
	"Routine maintenance, no issues found",
	"Replaced worn part during service",
	"Took longer than expected due to access",
	"Completed ahead of schedule",
	"Found minor issue, will monitor",
	"Used OEM replacement parts",
	"Cleaned thoroughly, good for another cycle",
	"Previous repair holding up well",
	"Adjusted per manufacturer specs",
	"Left notes for next service visit",
}
