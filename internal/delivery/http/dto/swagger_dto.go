package dto

type ResponseSuccess struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Success message"`
}

type AuthTokensResponseWrapper struct {
	ResponseSuccess
	Data AuthTokensResponse `json:"data"`
}

type AuthResponseWrapper struct {
	ResponseSuccess
	Data AuthResponse `json:"data"`
}

type UserResponseWrapper struct {
	ResponseSuccess
	Data UserResponse `json:"data"`
}

type AuthURLResponseWrapper struct {
	ResponseSuccess
	Data AuthURLResponse `json:"data"`
}

type MuzakkiResponseWrapper struct {
	ResponseSuccess
	Data MuzakkiResponse `json:"data"`
}

type MuzakkiListResponseWrapper struct {
	ResponseSuccess
	Data interface{} `json:"data"` // Contains pagination data
}

type AsnafResponseWrapper struct {
	ResponseSuccess
	Data AsnafResponse `json:"data"`
}

type AsnafListResponseWrapper struct {
	ResponseSuccess
	Data interface{} `json:"data"` // Contains pagination data
}

type MustahiqResponseWrapper struct {
	ResponseSuccess
	Data MustahiqResponse `json:"data"`
}

type MustahiqListResponseWrapper struct {
	ResponseSuccess
	Data interface{} `json:"data"` // Contains pagination data
}

type ProgramResponseWrapper struct {
	ResponseSuccess
	Data ProgramResponse `json:"data"`
}

type ProgramListResponseWrapper struct {
	ResponseSuccess
	Data interface{} `json:"data"` // Contains pagination data
}

type DonationReceiptResponseWrapper struct {
	ResponseSuccess
	Data DonationReceiptResponse `json:"data"`
}

type DonationReceiptListResponseWrapper struct {
	ResponseSuccess
	Data interface{} `json:"data"` // Contains pagination data
}

type DistributionResponseWrapper struct {
	ResponseSuccess
	Data DistributionResponse `json:"data"`
}

type DistributionListResponseWrapper struct {
	ResponseSuccess
	Data interface{} `json:"data"` // Contains pagination data
}

type ReportResponseWrapper struct {
	ResponseSuccess
	Data interface{} `json:"data"` // Generic for all reports
}

type ErrorResponseWrapper struct {
	Success bool        `json:"success" example:"false"`
	Message string      `json:"message" example:"Error message"`
	Errors  interface{} `json:"errors,omitempty"`
}
