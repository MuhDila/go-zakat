package dto

// Income Summary Response
type IncomeSummaryResponse struct {
	Period      string  `json:"period"`
	ZakatFitrah float64 `json:"zakat_fitrah"`
	ZakatMaal   float64 `json:"zakat_maal"`
	Infaq       float64 `json:"infaq"`
	Sadaqah     float64 `json:"sadaqah"`
	Total       float64 `json:"total"`
}

// Distribution Summary Responses
type DistributionSummaryByAsnafResponse struct {
	AsnafName        string  `json:"asnaf_name"`
	BeneficiaryCount int64   `json:"beneficiary_count"`
	TotalAmount      float64 `json:"total_amount"`
}

type DistributionSummaryByProgramResponse struct {
	ProgramName      string  `json:"program_name"`
	SourceFundType   string  `json:"source_fund_type"`
	BeneficiaryCount int64   `json:"beneficiary_count"`
	TotalAmount      float64 `json:"total_amount"`
}

// Fund Balance Response
type FundBalanceResponse struct {
	FundType string  `json:"fund_type"`
	TotalIn  float64 `json:"total_in"`
	TotalOut float64 `json:"total_out"`
	Balance  float64 `json:"balance"`
}

// Mustahiq History Response
type MustahiqHistoryItemResponse struct {
	DistributionDate string  `json:"distribution_date"`
	ProgramName      string  `json:"program_name"`
	SourceFundType   string  `json:"source_fund_type"`
	Amount           float64 `json:"amount"`
}

type MustahiqHistoryMustahiqInfo struct {
	ID        string `json:"id"`
	FullName  string `json:"full_name"`
	AsnafName string `json:"asnaf_name"`
	Address   string `json:"address"`
}

type MustahiqHistoryResponse struct {
	Mustahiq      MustahiqHistoryMustahiqInfo   `json:"mustahiq"`
	History       []MustahiqHistoryItemResponse `json:"history"`
	TotalReceived float64                       `json:"total_received"`
}
