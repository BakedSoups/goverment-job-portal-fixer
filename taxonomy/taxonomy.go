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
		{Name: "python", Label: "Python", Category: "Programming", Aliases: []string{"python", "python3", "py"}},
		{Name: "javascript", Label: "JavaScript", Category: "Programming", Aliases: []string{"javascript", "js", "node", "node.js", "typescript"}},
		{Name: "go", Label: "Go", Category: "Programming", Aliases: []string{"golang", "go language"}},
		{Name: "sql", Label: "SQL", Category: "Database", Aliases: []string{"sql", "structured query language", "query development"}},
		{Name: "postgresql", Label: "PostgreSQL", Category: "Database", Aliases: []string{"postgres", "postgresql", "postgres sql"}},
		{Name: "oracle", Label: "Oracle", Category: "Database", Aliases: []string{"oracle", "oracle c2m", "oracle database"}},
		{Name: "nosql", Label: "NoSQL", Category: "Database", Aliases: []string{"nosql", "no sql", "mongodb", "mongo", "document database", "non-relational"}},
		{Name: "data_analysis", Label: "Data Analysis", Category: "Data", Aliases: []string{"data analysis", "data analyst", "analytics", "reporting", "data quality", "dashboard"}},
		{Name: "business_intelligence", Label: "Business Intelligence", Category: "Data", Aliases: []string{"business intelligence", "bi", "tableau", "power bi", "cognos"}},
		{Name: "software_engineer", Label: "Software Engineer", Category: "Role", Aliases: []string{"software engineer", "software developer", "application developer", "programmer", "developer"}},
		{Name: "business_analyst", Label: "Business Analyst", Category: "Role", Aliases: []string{"business analyst", "systems analyst", "is business analyst"}},
		{Name: "cloud", Label: "Cloud", Category: "Infrastructure", Aliases: []string{"aws", "azure", "google cloud", "gcp", "cloud"}},
		{Name: "gis", Label: "GIS", Category: "Data", Aliases: []string{"gis", "geographic information system", "geospatial"}},
		{Name: "excel", Label: "Excel", Category: "Tools", Aliases: []string{"excel", "spreadsheet", "pivot table"}},
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
