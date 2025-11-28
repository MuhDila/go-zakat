package repository

// Result structs for reports
type IncomeSummaryResult struct {
	Period      string // YYYY-MM-DD or YYYY-MM depending on groupBy
	ZakatFitrah float64
	ZakatMaal   float64
	Infaq       float64
	Sadaqah     float64
	Total       float64
}

type DistributionSummaryByAsnafResult struct {
	AsnafName        string
	BeneficiaryCount int64
	TotalAmount      float64
}

type DistributionSummaryByProgramResult struct {
	ProgramName      string
	SourceFundType   string
	BeneficiaryCount int64
	TotalAmount      float64
}

type FundBalanceResult struct {
	FundType string
	TotalIn  float64
	TotalOut float64
	Balance  float64
}

type MustahiqHistoryItem struct {
	DistributionDate string
	ProgramName      string
	SourceFundType   string
	Amount           float64
}

type MustahiqHistoryResult struct {
	MustahiqID    string
	FullName      string
	AsnafName     string
	Address       string
	History       []MustahiqHistoryItem
	TotalReceived float64
}

type ReportRepository interface {
	GetIncomeSummary(dateFrom, dateTo, groupBy string) ([]IncomeSummaryResult, error)
	GetDistributionSummary(dateFrom, dateTo, groupBy, sourceFundType string) (interface{}, error)
	GetFundBalance(dateFrom, dateTo string) ([]FundBalanceResult, error)
	GetMustahiqHistory(mustahiqID string) (*MustahiqHistoryResult, error)
}
