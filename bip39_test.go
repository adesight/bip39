package bip39_test

import (
	"fmt"
	"testing"

	"github.com/adesight/bip39"
)

func TestValidateMnemonic(t *testing.T) {
	type args struct {
		mnemonic string
		lang     bip39.Language
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "English",
			args: args{
				mnemonic: "check fiscal fit sword unlock rough lottery tool sting pluck bulb random",
				lang:     bip39.English,
			},
			want: true,
		},
		{
			name: "Englishx2",
			args: args{
				mnemonic: "rich soon pool legal busy add couch tower goose security raven anger",
				lang:     bip39.English,
			},
			want: true,
		},
		{
			name: "EnglishValidLength",
			args: args{
				mnemonic: "rich soon pool legal busy add couch tower goose security raven",
				lang:     bip39.English,
			},
			want: false,
		},
		{
			name: "EnglishNoWord",
			args: args{
				mnemonic: "rich soon pool legal busy add couch tower goose security women",
				lang:     bip39.English,
			},
			want: false,
		},
		{
			name: "EnglishChecksumError",
			args: args{
				mnemonic: "rich soon pool legal busy add couch tower goose security base",
				lang:     bip39.English,
			},
			want: false,
		},
		{
			name: "ChineseSimplified",
			args: args{
				mnemonic: "氮 冠 锋 枪 做 到 容 枯 获 槽 弧 部",
				lang:     bip39.ChineseSimplified,
			},
			want: true,
		},
		{
			name: "ChineseTraditional",
			args: args{
				mnemonic: "氮 冠 鋒 槍 做 到 容 枯 獲 槽 弧 部",
				lang:     bip39.ChineseTraditional,
			},
			want: true,
		},
		{
			name: "Japanese",
			args: args{
				mnemonic: "ねほりはほり　ひらがな　とさか　そつう　おうじ　あてな　きくらげ　みもと　してつ　ぱそこん　にってい　いこつ",
				lang:     bip39.Japanese,
			},
			want: true,
		},
		{
			name: "Spanish",
			args: args{
				mnemonic: "posible ruptura ozono ligero bobina acto chuleta tetera gol realidad pez alerta",
				lang:     bip39.Spanish,
			},
			want: true,
		},
		{
			name: "French",
			args: args{
				mnemonic: "pieuvre revivre nuptial implorer blinder accroche chute syntaxe félin promener parcelle aimable",
				lang:     bip39.French,
			},
			want: true,
		},
		{
			name: "Italian",
			args: args{
				mnemonic: "risultato siccome prenotare mimosa bosco adottare continuo tifare ignaro sbloccato residente alticcio",
				lang:     bip39.Italian,
			},
			want: true,
		},
		{
			name: "Korean",
			args: args{
				mnemonic: "전망 차선 이전 실장 기간 간판 대접 판단 생명 존재 잠깐 건축",
				lang:     bip39.Korean,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bip39.ValidateMnemonic(tt.args.mnemonic, tt.args.lang); got != tt.want {
				t.Errorf("ValidateMnemonic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleValidateMnemonic() {
	var mnemonic = "check fiscal fit sword unlock rough lottery tool sting pluck bulb random"
	fmt.Println(bip39.ValidateMnemonic(mnemonic, bip39.English))

	// Output:
	// true
}
