package taxonomy

import "strings"

type Concept struct {
	Name     string
	Label    string
	Category string
	Aliases  []string
}

func Concepts() []Concept {
	return []Concept{
		{Name: "data_analytics", Label: "Data Analytics", Category: "Data & Analytics", Aliases: []string{"business analytics", "data analysis", "data analyst", "analytics", "reporting", "data quality", "dashboard", "performance analyst", "people science analyst", "user research analyst"}},
		{Name: "business_analysis", Label: "Business Analysis", Category: "Data & Analytics", Aliases: []string{"business analyst", "business analytics", "systems analyst", "is business analyst", "requirements analysis", "process improvement", "stakeholder requirements"}},
		{Name: "business_intelligence", Label: "Business Intelligence", Category: "Data & Analytics", Aliases: []string{"business intelligence", "bi", "tableau", "power bi", "cognos", "dashboarding", "metrics"}},
		{Name: "sql", Label: "SQL", Category: "Data & Analytics", Aliases: []string{"sql", "structured query language", "query development", "database query", "queries"}},
		{Name: "excel", Label: "Excel", Category: "Data & Analytics", Aliases: []string{"excel", "spreadsheet", "pivot table", "pivot tables", "vlookup", "workbook"}},
		{Name: "gis", Label: "GIS", Category: "Data & Analytics", Aliases: []string{"gis", "geographic information system", "geospatial", "mapping", "surveying and mapping"}},

		{Name: "software_engineering", Label: "Software Engineering", Category: "IT & Software", Aliases: []string{"software engineer", "software developer", "application developer", "programmer", "developer", "principal system integration engineer", "system integration engineer"}},
		{Name: "python", Label: "Python", Category: "IT & Software", Aliases: []string{"python", "python3", "py"}},
		{Name: "javascript", Label: "JavaScript", Category: "IT & Software", Aliases: []string{"javascript", "js", "node", "node.js", "typescript"}},
		{Name: "oracle", Label: "Oracle", Category: "IT & Software", Aliases: []string{"oracle", "oracle c2m", "oracle database", "oracle apex", "oracle bpel", "oracle soa"}},
		{Name: "cloud_infrastructure", Label: "Cloud & Infrastructure", Category: "IT & Software", Aliases: []string{"aws", "azure", "google cloud", "gcp", "cloud", "infrastructure", "disaster recovery", "resilience", "enterprise systems"}},
		{Name: "enterprise_architecture", Label: "Enterprise Architecture", Category: "IT & Software", Aliases: []string{"enterprise architect", "solutions architect", "solution architect", "architecture", "systems architecture"}},
		{Name: "project_management", Label: "Project Management", Category: "IT & Software", Aliases: []string{"project manager", "project management", "it project manager", "modernization project", "project lead"}},

		{Name: "civil_engineering", Label: "Civil Engineering", Category: "Engineering & Infrastructure", Aliases: []string{"civil engineer", "resident engineer", "assistant engineer", "engineering associate", "engineer architect", "capital project", "construction management"}},
		{Name: "electrical_engineering", Label: "Electrical Engineering", Category: "Engineering & Infrastructure", Aliases: []string{"electrical engineer", "signal engineer", "signal/electrical", "electrical assistant engineer", "power engineering", "transmission and distribution"}},
		{Name: "utilities_water_power", Label: "Utilities, Water & Power", Category: "Engineering & Infrastructure", Aliases: []string{"public utilities", "water division", "water resources", "power enterprise", "wastewater", "sewage plant", "utility specialist", "energy specialist"}},
		{Name: "surveying_mapping", Label: "Surveying & Mapping", Category: "Engineering & Infrastructure", Aliases: []string{"surveyor", "surveying", "mapping", "land survey", "chief surveyor"}},

		{Name: "facilities_maintenance", Label: "Facilities Maintenance", Category: "Facilities & Trades", Aliases: []string{"facilities", "maintenance", "maintenance planner", "stationary engineer", "custodian", "building maintenance"}},
		{Name: "electrician", Label: "Electrician", Category: "Facilities & Trades", Aliases: []string{"electrician", "utility electrician", "electrical worker", "line worker", "transmission line worker"}},
		{Name: "mechanic_trades", Label: "Mechanics & Skilled Trades", Category: "Facilities & Trades", Aliases: []string{"mechanic", "automotive machinist", "heavy duty mechanic", "cement mason", "asphalt finisher", "trades"}},

		{Name: "nursing", Label: "Nursing", Category: "Health & Clinical", Aliases: []string{"nurse", "registered nurse", "public health nurse", "licensed vocational nurse", "nurse manager", "nursing assistant", "certified nursing assistant", "cna", "lvn", "rn"}},
		{Name: "behavioral_health", Label: "Behavioral Health", Category: "Health & Clinical", Aliases: []string{"behavioral health", "mental health", "psychologist", "psychiatric", "clinician", "therapy", "counseling"}},
		{Name: "clinical_care", Label: "Clinical Care", Category: "Health & Clinical", Aliases: []string{"clinical", "patient care", "medical record", "medical records", "radiologic technologist", "surgical procedures", "occupational therapist", "physical therapist", "health worker"}},
		{Name: "pharmacy", Label: "Pharmacy", Category: "Health & Clinical", Aliases: []string{"pharmacy", "pharmacist", "clinical pharmacist", "pharmacy technician"}},
		{Name: "public_health", Label: "Public Health", Category: "Health & Clinical", Aliases: []string{"public health", "environmental health", "health inspector", "industrial hygienist", "consumer protection", "jail health"}},

		{Name: "management_leadership", Label: "Management & Leadership", Category: "Management & Programs", Aliases: []string{"management", "manager", "leadership", "nurse manager", "budget director", "director of enforcement", "director of safety", "director of public affairs", "physician director", "deputy director", "managing director", "bureau manager", "team leader", "supervising pharmacist", "chief of real property", "chief surveyor"}},
		{Name: "program_operations", Label: "Program Operations", Category: "Management & Programs", Aliases: []string{"programs", "operations", "integrated operations", "program operations", "service delivery", "operations manager", "director for programs"}},
		{Name: "human_resources", Label: "Human Resources", Category: "Management & Programs", Aliases: []string{"human resources", "citywide leaves", "accommodation manager", "people science", "employee relations"}},

		{Name: "public_safety", Label: "Public Safety", Category: "Public Safety", Aliases: []string{"police", "firefighter", "fire department", "public safety", "institutional police", "community police", "crossing guard", "background investigator"}},
		{Name: "probation_corrections", Label: "Probation & Corrections", Category: "Public Safety", Aliases: []string{"juvenile probation", "juvenile hall", "juvenile hall counselor", "corrections", "jail"}},

		{Name: "legal", Label: "Legal", Category: "Legal & Finance", Aliases: []string{"attorney", "city attorney", "district attorney", "law librarian", "legal", "civil/criminal", "real estate and finance"}},
		{Name: "finance_budget", Label: "Finance & Budget", Category: "Legal & Finance", Aliases: []string{"finance", "budget", "budget director", "public finance", "treasurer", "tax collector", "patient accounts", "real estate and finance"}},
		{Name: "real_estate_property", Label: "Real Estate & Property", Category: "Legal & Finance", Aliases: []string{"real estate", "real property", "assessor", "recorder", "property"}},

		{Name: "administration", Label: "Administration", Category: "Admin & Customer Service", Aliases: []string{"administration", "administrative", "administrative analyst", "public service aide", "assistant to professionals", "associate to professionals", "medical record technician", "registrar"}},
		{Name: "communications_public_affairs", Label: "Communications & Public Affairs", Category: "Admin & Customer Service", Aliases: []string{"communications", "public affairs", "media relations", "spokesperson", "external affairs"}},
		{Name: "customer_service", Label: "Customer Service", Category: "Admin & Customer Service", Aliases: []string{"customer service", "front desk", "community service", "service aide"}},

		{Name: "recreation_parks", Label: "Recreation & Parks", Category: "Recreation & Service", Aliases: []string{"recreation", "parks", "camp mather", "lifeguard", "pool lifeguard", "recreation leader"}},
		{Name: "food_service", Label: "Food Service", Category: "Recreation & Service", Aliases: []string{"food service", "cook", "assistant cook", "dietetic", "dietetic technician"}},
	}
}

func AliasMap() map[string]Concept {
	out := make(map[string]Concept)
	for _, concept := range Concepts() {
		out[concept.Name] = concept
		out[strings.ToLower(concept.Label)] = concept
		for _, alias := range concept.Aliases {
			out[strings.ToLower(alias)] = concept
		}
	}
	return out
}
