package weex

type WeexBaseResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

type GetOpenOrderListRequest struct {
	TraderUserID string `json:"traderUserId"`
	PageNo       int    `json:"pageNo"`
	PageSize     int    `json:"pageSize"`
	ContractId   string `json:"contractId"`
	LanguageType int    `json:"languageType"`
}

type GetOpenOrderListResponse struct {
	WeexBaseResponse
	Data struct {
		NextFlag bool        `json:"nextFlag"`
		Totals   int         `json:"totals"`
		Rows     []OpenOrder `json:"rows"`
	}
}

type GetMetaDataV2Response struct {
	WeexBaseResponse
	Data struct {
		ContractList []struct {
			// 交易对代码
			Ci string `json:"ci"`
			// 交易对名称
			Cn string `json:"cn"`
		} `json:"contractList"`
	}
}

type OpenOrder struct {
	ID                    string `json:"id"`
	TraderAccountID       string `json:"traderAccountId"`
	AccountID             string `json:"accountId"`
	ContractID            string `json:"contractId"`
	TrackingType          string `json:"trackingType"`
	OpenOrderID           string `json:"openOrderId"`
	ParentTrackingOrderID string `json:"parentTrackingOrderId"`
	PositionSide          string `json:"positionSide"`
	OpenLeverage          string `json:"openLeverage"`
	OpenSize              string `json:"openSize"`
	OpenDone              bool   `json:"openDone"`
	AverageOpenPrice      string `json:"averageOpenPrice"`
	OpenFillSize          string `json:"openFillSize"`
	OpenMarginAmount      string `json:"openMarginAmount"`
	OpenFee               string `json:"openFee"`
	OpenTime              string `json:"openTime"`
	LiquidatePrice        string `json:"liquidatePrice"`
	BankruptPrice         string `json:"bankruptPrice"`
	AverageClosePrice     string `json:"averageClosePrice"`
	CloseFillSize         string `json:"closeFillSize"`
	CloseTime             string `json:"closeTime"`
	CloseOrderID          string `json:"closeOrderId"`
	RealizedPnl           string `json:"realizedPnl"`
	CloseFee              string `json:"closeFee"`
	FundingFee            string `json:"fundingFee"`
	ProfitsShareRatio     string `json:"profitsShareRatio"`
	CumShareAmount        string `json:"cumShareAmount"`
	ShareProfit           string `json:"shareProfit"`
	ExecuteShareProfit    string `json:"executeShareProfit"`
	ActualShareProfit     string `json:"actualShareProfit"`
	TakeProfitPrice       string `json:"takeProfitPrice"`
	TakeProfitTrigger     bool   `json:"takeProfitTrigger"`
	StopLossPrice         string `json:"stopLossPrice"`
	StopLossTrigger       bool   `json:"stopLossTrigger"`
	Status                string `json:"status"`
	CreatedTime           string `json:"createdTime"`
	UpdatedTime           string `json:"updatedTime"`
	ExtraDataJSON         string `json:"extraDataJson"`
	ClosedBy              string `json:"closedBy"`
	UserID                string `json:"userId"`
	TradeUserID           string `json:"tradeUserId"`
	TraderName            string `json:"traderName"`
	TakeProfitOrderID     string `json:"takeProfitOrderId"`
	StopLossOrderID       string `json:"stopLossOrderId"`
	SeparatedMode         string `json:"separatedMode"`
	MarginMode            string `json:"marginMode"`
	ProfitRate            string `json:"profitRate"`
	ProfitRateText        string `json:"profitRateText"`
	NetProfit             string `json:"netProfit"`
}
