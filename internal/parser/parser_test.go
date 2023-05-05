package parser

import (
	"github.com/goccy/go-json"
	tronApi "github.com/kattana-io/tron-objects-api/pkg/api"
	"testing"
)

func Test_getMethodId(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Transfer event",
			args: args{
				input: "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
			},
			want: 0xddf252ad,
		},
		{
			name: "Token purchase event",
			args: args{
				input: "cd60aa75dea3072fbc07ae6d7d856b5dc5f4eee88854f5b4abf7b680ef8bc50f",
			},
			want: 0xcd60aa75,
		},
		{
			name: "Snapshot event",
			args: args{
				input: "cc7244d3535e7639366f8c5211527112e01de3ec7449ee3a6e66b007f4065a70",
			},
			want: 0xcc7244d3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMethodID(tt.args.input); got != tt.want {
				t.Errorf("getMethodId() = %v, want %v", got, tt.want)
			}
		})
	}
}

const RevertTxExample = `{
            "ret": [
                {
                    "contractRet": "REVERT"
                }
            ],
            "signature": [
                "35937d9eb2ee2a518b216efeb196d3413e1827ec4821fd883d9de9b006e1634540815b479c5ee66f71b01db92e185dd2826ac13531165226d4958189516ebef900"
            ],
            "txID": "c3c4dbeeb918e721fe252656cfb0d8e0019db647e9b20d286e2194c219e9cbbc",
            "raw_data": {
                "contract": [
                    {
                        "parameter": {
                            "value": {
                                "data": "a3082be9000000000000000000000000000000000000000000000000000000000000005f0000000000000000000000000000000000000000000000000000000000000000",
                                "owner_address": "4103fd756bc58f6d9ad87a406515f5d544c0dca2f2",
                                "contract_address": "412ec5f63da00583085d4c2c5e8ec3c8d17bde5e28",
                                "call_value": 50000000
                            },
                            "type_url": "type.googleapis.com/protocol.TriggerSmartContract"
                        },
                        "type": "TriggerSmartContract"
                    }
                ],
                "ref_block_bytes": "6913",
                "ref_block_hash": "6cc41a5a347e14b0",
                "expiration": 1552011786000,
                "fee_limit": 6000000,
                "timestamp": 1552011728175
            },
            "raw_data_hex": "0a02691322086cc41a5a347e14b04090eef1d8952d5ab301081f12ae010a31747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e54726967676572536d617274436f6e747261637412790a154103fd756bc58f6d9ad87a406515f5d544c0dca2f21215412ec5f63da00583085d4c2c5e8ec3c8d17bde5e281880e1eb172244a3082be9000000000000000000000000000000000000000000000000000000000000005f000000000000000000000000000000000000000000000000000000000000000070afaaeed8952d9001809bee02"
        },`
const TransferTxExample = ` {
            "ret": [
                {
                    "contractRet": "SUCCESS"
                }
            ],
            "signature": [
                "ae848bd11900e544ef868044e80ef58d9274d7c193c65ab87ff3e40d48e570aa4bb0ceeedaa33f6e32224ae80a6a12da5bba29d3e6e9ae62794be5a5cc9eebfd01"
            ],
            "txID": "5e5a18bb6bca8ab35975d4adba6a6cb07bba812a5f41c0fdd7815feb2aa13bca",
            "raw_data": {
                "data": "e4b89ce696b9e6b187e58cbae59d97e993bee6b8b8e6888fe8bdace8b4a62ce69c9fe58fb73a323032322d30382d3031203631382c207369676e3a3438633466",
                "contract": [
                    {
                        "parameter": {
                            "value": {
                                "amount": 618000,
                                "owner_address": "415e56893a570bd2866a085d789c14e9f8a70db155",
                                "to_address": "419f9ad4eb70c9b09686203901b205d1946eac6a6d"
                            },
                            "type_url": "type.googleapis.com/protocol.TransferContract"
                        },
                        "type": "TransferContract"
                    }
                ],
                "ref_block_bytes": "ec67",
                "ref_block_hash": "a3690fe0b0b7b26b",
                "expiration": 1659338868000,
                "timestamp": 1659338810483
            },
            "raw_data_hex": "0a02ec672208a3690fe0b0b7b26b40a08ab7c2a5305240e4b89ce696b9e6b187e58cbae59d97e993bee6b8b8e6888fe8bdace8b4a62ce69c9fe58fb73a323032322d30382d3031203631382c207369676e3a34386334665a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a15415e56893a570bd2866a085d789c14e9f8a70db1551215419f9ad4eb70c9b09686203901b205d1946eac6a6d1890dc2570f3c8b3c2a530"
        },`
const TradeTxExample = `{
            "ret": [
                {
                    "contractRet": "SUCCESS"
                }
            ],
            "signature": [
                "9070771a4128e3409971f5b3d75166dde6f1b19eb3e288ff1d48cb1eb0849f93fa084baf0a732448e3c7cf220999adf9a6b0bf4b07d659139909f0364e54fdea01"
            ],
            "txID": "9f9f64c995d71bf2f3cd1b49a5578f29efd93946d7b2cb74a9764b2b47aed4a5",
            "raw_data": {
                "contract": [
                    {
                        "parameter": {
                            "value": {
                                "data": "ddf7e1a700000000000000000000000000000000000000000000000009bb678eea7aab180000000000000000000000000000000000000000000000000000000006dfaec700000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000062e78069000000000000000000000000a614f803b6fd780986a42c78ec9c7f77e6ded13c",
                                "owner_address": "41b7fc9f70a161c1800e38a152c4adabf9bb2cedce",
                                "contract_address": "4146cd4764d54941a937f7eeb6d9b8ed3e7280446d"
                            },
                            "type_url": "type.googleapis.com/protocol.TriggerSmartContract"
                        },
                        "type": "TriggerSmartContract"
                    }
                ],
                "ref_block_bytes": "ec75",
                "ref_block_hash": "4596592087a6cb86",
                "expiration": 1659338856000,
                "fee_limit": 100000000,
                "timestamp": 1659338796726
            },
            "raw_data_hex": "0a02ec7522084596592087a6cb8640c0acb6c2a5305a9002081f128b020a31747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e54726967676572536d617274436f6e747261637412d5010a1541b7fc9f70a161c1800e38a152c4adabf9bb2cedce12154146cd4764d54941a937f7eeb6d9b8ed3e7280446d22a401ddf7e1a700000000000000000000000000000000000000000000000009bb678eea7aab180000000000000000000000000000000000000000000000000000000006dfaec700000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000062e78069000000000000000000000000a614f803b6fd780986a42c78ec9c7f77e6ded13c70b6ddb2c2a530900180c2d72f"
        }`

func CreateValidTradeTx() *tronApi.Transaction {
	data := tronApi.Transaction{}
	json.Unmarshal([]byte(TradeTxExample), &data)
	return &data
}

func CreateValidTransferTx() *tronApi.Transaction {
	data := tronApi.Transaction{}
	json.Unmarshal([]byte(TransferTxExample), &data)
	return &data
}

func CreateRevertTx() *tronApi.Transaction {
	data := tronApi.Transaction{}
	json.Unmarshal([]byte(RevertTxExample), &data)
	return &data
}

func Test_isSuccessCall(t *testing.T) {
	type args struct {
		transaction *tronApi.Transaction
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Should pass valid trade tx",
			args: args{
				transaction: CreateValidTradeTx(),
			},
			want: true,
		},
		{
			name: "Should pass valid transfer tx",
			args: args{
				transaction: CreateValidTransferTx(),
			},
			want: true,
		},
		{
			name: "Should not pass revert tx",
			args: args{
				transaction: CreateRevertTx(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSuccessCall(tt.args.transaction); got != tt.want {
				t.Errorf("isSuccessCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isNotTransferCall(t *testing.T) {
	type args struct {
		transaction *tronApi.Transaction
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Should pass valid tx",
			args: args{
				transaction: CreateValidTradeTx(),
			},
			want: true,
		},
		{
			name: "Should not pass valid transfer tx",
			args: args{
				transaction: CreateValidTransferTx(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNotTransferCall(tt.args.transaction); got != tt.want {
				t.Errorf("isNotTransferCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasContractCalls(t *testing.T) {
	type args struct {
		transaction *tronApi.Transaction
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Should pass valid trader tx",
			args: args{
				transaction: CreateValidTradeTx(),
			},
			want: true,
		},
		{
			name: "Should pass valid transfer tx",
			args: args{
				transaction: CreateValidTransferTx(),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasContractCalls(tt.args.transaction); got != tt.want {
				t.Errorf("hasContractCalls() = %v, want %v", got, tt.want)
			}
		})
	}
}
