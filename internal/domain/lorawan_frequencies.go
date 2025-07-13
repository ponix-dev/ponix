package domain

// LoRaWANFrequencyPlan represents a LoRaWAN frequency plan
type LoRaWANFrequencyPlan string

// LoRaWAN frequency plans as defined by the LoRa Alliance
const (
	// Europe
	FreqPlanEU863_870 LoRaWANFrequencyPlan = "EU_863_870"
	FreqPlanEU433     LoRaWANFrequencyPlan = "EU_433"

	// North America
	FreqPlanUS902_928 LoRaWANFrequencyPlan = "US_902_928"

	// Asia-Pacific
	FreqPlanAS923     LoRaWANFrequencyPlan = "AS_923"
	FreqPlanAS923_2   LoRaWANFrequencyPlan = "AS_923_2"
	FreqPlanAS923_3   LoRaWANFrequencyPlan = "AS_923_3"
	FreqPlanAS923_4   LoRaWANFrequencyPlan = "AS_923_4"

	// Australia
	FreqPlanAU915_928 LoRaWANFrequencyPlan = "AU_915_928"

	// China
	FreqPlanCN470_510 LoRaWANFrequencyPlan = "CN_470_510"
	FreqPlanCN779_787 LoRaWANFrequencyPlan = "CN_779_787"

	// India
	FreqPlanIN865_867 LoRaWANFrequencyPlan = "IN_865_867"

	// Korea
	FreqPlanKR920_923 LoRaWANFrequencyPlan = "KR_920_923"

	// Russia
	FreqPlanRU864_870 LoRaWANFrequencyPlan = "RU_864_870"
)

// String returns the string representation of the frequency plan
func (f LoRaWANFrequencyPlan) String() string {
	return string(f)
}

// GetFrequencyPlansByRegion returns available frequency plans for a given region
func GetFrequencyPlansByRegion(region string) []LoRaWANFrequencyPlan {
	switch region {
	case "EU", "Europe":
		return []LoRaWANFrequencyPlan{
			FreqPlanEU863_870,
			FreqPlanEU433,
		}
	case "US", "North America", "NA":
		return []LoRaWANFrequencyPlan{
			FreqPlanUS902_928,
		}
	case "AS", "Asia", "Asia-Pacific":
		return []LoRaWANFrequencyPlan{
			FreqPlanAS923,
			FreqPlanAS923_2,
			FreqPlanAS923_3,
			FreqPlanAS923_4,
		}
	case "AU", "Australia":
		return []LoRaWANFrequencyPlan{
			FreqPlanAU915_928,
		}
	case "CN", "China":
		return []LoRaWANFrequencyPlan{
			FreqPlanCN470_510,
			FreqPlanCN779_787,
		}
	case "IN", "India":
		return []LoRaWANFrequencyPlan{
			FreqPlanIN865_867,
		}
	case "KR", "Korea":
		return []LoRaWANFrequencyPlan{
			FreqPlanKR920_923,
		}
	case "RU", "Russia":
		return []LoRaWANFrequencyPlan{
			FreqPlanRU864_870,
		}
	default:
		return []LoRaWANFrequencyPlan{
			FreqPlanUS902_928, // Default to US plan
		}
	}
}

// GetAllFrequencyPlans returns all available frequency plans
func GetAllFrequencyPlans() []LoRaWANFrequencyPlan {
	return []LoRaWANFrequencyPlan{
		FreqPlanEU863_870,
		FreqPlanEU433,
		FreqPlanUS902_928,
		FreqPlanAS923,
		FreqPlanAS923_2,
		FreqPlanAS923_3,
		FreqPlanAS923_4,
		FreqPlanAU915_928,
		FreqPlanCN470_510,
		FreqPlanCN779_787,
		FreqPlanIN865_867,
		FreqPlanKR920_923,
		FreqPlanRU864_870,
	}
}

// IsValidFrequencyPlan checks if a frequency plan is valid
func IsValidFrequencyPlan(plan string) bool {
	allPlans := GetAllFrequencyPlans()
	for _, validPlan := range allPlans {
		if validPlan.String() == plan {
			return true
		}
	}
	return false
}

// GetDefaultFrequencyPlan returns the default frequency plan for a region
func GetDefaultFrequencyPlan(region string) LoRaWANFrequencyPlan {
	plans := GetFrequencyPlansByRegion(region)
	if len(plans) > 0 {
		return plans[0]
	}
	return FreqPlanUS902_928 // Fallback to US plan
}

// FrequencyPlanInfo contains metadata about a frequency plan
type FrequencyPlanInfo struct {
	ID          LoRaWANFrequencyPlan
	Name        string
	Description string
	Region      string
	Frequency   string
	Bandwidth   string
}

// GetFrequencyPlanInfo returns detailed information about frequency plans
func GetFrequencyPlanInfo() map[LoRaWANFrequencyPlan]FrequencyPlanInfo {
	return map[LoRaWANFrequencyPlan]FrequencyPlanInfo{
		FreqPlanEU863_870: {
			ID:          FreqPlanEU863_870,
			Name:        "EU 863-870 MHz",
			Description: "European LoRaWAN frequency plan",
			Region:      "EU",
			Frequency:   "863-870 MHz",
			Bandwidth:   "125/250 kHz",
		},
		FreqPlanEU433: {
			ID:          FreqPlanEU433,
			Name:        "EU 433 MHz",
			Description: "European 433 MHz LoRaWAN frequency plan",
			Region:      "EU",
			Frequency:   "433 MHz",
			Bandwidth:   "125 kHz",
		},
		FreqPlanUS902_928: {
			ID:          FreqPlanUS902_928,
			Name:        "US 902-928 MHz",
			Description: "North America LoRaWAN frequency plan",
			Region:      "US",
			Frequency:   "902-928 MHz",
			Bandwidth:   "125/500 kHz",
		},
		FreqPlanAS923: {
			ID:          FreqPlanAS923,
			Name:        "AS 923 MHz",
			Description: "Asia-Pacific LoRaWAN frequency plan",
			Region:      "AS",
			Frequency:   "923 MHz",
			Bandwidth:   "125/250 kHz",
		},
		FreqPlanAS923_2: {
			ID:          FreqPlanAS923_2,
			Name:        "AS 923-2 MHz",
			Description: "Asia-Pacific LoRaWAN frequency plan (variant 2)",
			Region:      "AS",
			Frequency:   "923 MHz",
			Bandwidth:   "125/250 kHz",
		},
		FreqPlanAS923_3: {
			ID:          FreqPlanAS923_3,
			Name:        "AS 923-3 MHz",
			Description: "Asia-Pacific LoRaWAN frequency plan (variant 3)",
			Region:      "AS",
			Frequency:   "923 MHz",
			Bandwidth:   "125/250 kHz",
		},
		FreqPlanAS923_4: {
			ID:          FreqPlanAS923_4,
			Name:        "AS 923-4 MHz",
			Description: "Asia-Pacific LoRaWAN frequency plan (variant 4)",
			Region:      "AS",
			Frequency:   "923 MHz",
			Bandwidth:   "125/250 kHz",
		},
		FreqPlanAU915_928: {
			ID:          FreqPlanAU915_928,
			Name:        "AU 915-928 MHz",
			Description: "Australia LoRaWAN frequency plan",
			Region:      "AU",
			Frequency:   "915-928 MHz",
			Bandwidth:   "125/500 kHz",
		},
		FreqPlanCN470_510: {
			ID:          FreqPlanCN470_510,
			Name:        "CN 470-510 MHz",
			Description: "China LoRaWAN frequency plan",
			Region:      "CN",
			Frequency:   "470-510 MHz",
			Bandwidth:   "125/250 kHz",
		},
		FreqPlanCN779_787: {
			ID:          FreqPlanCN779_787,
			Name:        "CN 779-787 MHz",
			Description: "China 779-787 MHz LoRaWAN frequency plan",
			Region:      "CN",
			Frequency:   "779-787 MHz",
			Bandwidth:   "125/250 kHz",
		},
		FreqPlanIN865_867: {
			ID:          FreqPlanIN865_867,
			Name:        "IN 865-867 MHz",
			Description: "India LoRaWAN frequency plan",
			Region:      "IN",
			Frequency:   "865-867 MHz",
			Bandwidth:   "125/250 kHz",
		},
		FreqPlanKR920_923: {
			ID:          FreqPlanKR920_923,
			Name:        "KR 920-923 MHz",
			Description: "Korea LoRaWAN frequency plan",
			Region:      "KR",
			Frequency:   "920-923 MHz",
			Bandwidth:   "125/250 kHz",
		},
		FreqPlanRU864_870: {
			ID:          FreqPlanRU864_870,
			Name:        "RU 864-870 MHz",
			Description: "Russia LoRaWAN frequency plan",
			Region:      "RU",
			Frequency:   "864-870 MHz",
			Bandwidth:   "125/250 kHz",
		},
	}
}