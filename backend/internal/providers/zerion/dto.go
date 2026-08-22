package zerion

type positionsResponse struct {
	Data []positionResource `json:"data"`
}

type positionResource struct {
	ID         string            `json:"id"`
	Attributes positionAttributes `json:"attributes"`
}

type positionAttributes struct {
	Quantity     quantityDTO   `json:"quantity"`
	FungibleInfo fungibleInfo  `json:"fungible_info"`
	UpdatedAt    string        `json:"updated_at"`
	PositionType string        `json:"position_type"`
}

type transactionsResponse struct {
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
	Data []transactionResource `json:"data"`
}

type transactionResource struct {
	ID         string               `json:"id"`
	Attributes transactionAttributes `json:"attributes"`
}

type transactionAttributes struct {
	Hash         string        `json:"hash"`
	MinedAtBlock int           `json:"mined_at_block"`
	MinedAt      string        `json:"mined_at"`
	SentFrom     string        `json:"sent_from"`
	SentTo       string        `json:"sent_to"`
	Status       string        `json:"status"`
	OperationType string       `json:"operation_type"`
	ApplicationMetadata *struct {
		Name string `json:"name"`
	} `json:"application_metadata"`
	Transfers    []transferDTO `json:"transfers"`
}

type transferDTO struct {
	Direction    string       `json:"direction"`
	Quantity     quantityDTO  `json:"quantity"`
	FungibleInfo fungibleInfo `json:"fungible_info"`
}

type quantityDTO struct {
	Int      string `json:"int"`
	Decimals int    `json:"decimals"`
	Numeric  string `json:"numeric"`
}

type fungibleInfo struct {
	Name           string              `json:"name"`
	Symbol         string              `json:"symbol"`
	Icon           *iconDTO            `json:"icon"`
	Implementations []implementationDTO `json:"implementations"`
}

type iconDTO struct {
	URL string `json:"url"`
}

type implementationDTO struct {
	ChainID  string  `json:"chain_id"`
	Address  *string `json:"address"`
	Decimals int     `json:"decimals"`
}
