package scraper

type ListResponse struct {
	Offset     int       `json:"offset"`
	Limit      int       `json:"limit"`
	TotalFound int       `json:"totalFound"`
	Content    []Summary `json:"content"`
}

type Summary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Ref       string `json:"ref"`
	RefNumber string `json:"refNumber"`
}

type Source struct {
	ID         string
	Name       string
	Kind       string
	Identifier string
	Region     string
}

type Posting struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	RefNumber    string        `json:"refNumber"`
	ReleasedDate string        `json:"releasedDate"`
	PostingURL   string        `json:"postingUrl"`
	ApplyURL     string        `json:"applyUrl"`
	Location     Location      `json:"location"`
	Department   Label         `json:"department"`
	Employment   Label         `json:"typeOfEmployment"`
	Experience   Label         `json:"experienceLevel"`
	CustomField  []CustomField `json:"customField"`
	JobAd        JobAd         `json:"jobAd"`
	Source       Source        `json:"-"`
}

type Location struct {
	City         string `json:"city"`
	Region       string `json:"region"`
	Country      string `json:"country"`
	FullLocation string `json:"fullLocation"`
	Remote       bool   `json:"remote"`
	Hybrid       bool   `json:"hybrid"`
}

type Label struct {
	ID    any    `json:"id"`
	Label string `json:"label"`
}

type CustomField struct {
	FieldLabel string `json:"fieldLabel"`
	ValueLabel string `json:"valueLabel"`
}

type JobAd struct {
	Sections Sections `json:"sections"`
}

type Sections struct {
	CompanyDescription    Section `json:"companyDescription"`
	JobDescription        Section `json:"jobDescription"`
	Qualifications        Section `json:"qualifications"`
	AdditionalInformation Section `json:"additionalInformation"`
}

type Section struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}
