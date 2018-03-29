package util

import (
	refmath "math"
	"math/big"
	"regexp"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

var Ether = math.BigPow(10, 18)
var Shannon = math.BigPow(10, 9)

var pow256 = math.BigPow(2, 256)
var addressPattern = regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
var zeroHash = regexp.MustCompile("^0?x?0+$")

func IsValidHexAddress(s string) bool {
	if IsZeroHash(s) || !addressPattern.MatchString(s) {
		return false
	}
	return true
}

func IsZeroHash(s string) bool {
	return zeroHash.MatchString(s)
}

func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GetTargetHex(diff int64) string {
	difficulty := big.NewInt(diff)
	diff1 := new(big.Int).Div(pow256, difficulty)
	return string(common.ToHex(diff1.Bytes()))
}

func TargetHexToDiff(targetHex string) *big.Int {
	targetBytes := common.FromHex(targetHex)
	return new(big.Int).Div(pow256, new(big.Int).SetBytes(targetBytes))
}

func ToHex(n int64) string {
	return "0x0" + strconv.FormatInt(n, 16)
}

func FormatReward(reward *big.Int) string {
	return reward.String()
}

func FormatRatReward(reward *big.Rat) string {
	wei := new(big.Rat).SetInt(Ether)
	reward = reward.Quo(reward, wei)
	return reward.FloatString(8)
}

func GetPPSRate(shareDiff, netDiff int64, fee float64) float64 {
	// Calculate current PPS rate
	wei := new(big.Rat).SetInt(Ether)
	wei.Mul(wei, new(big.Rat).SetInt64(3))
	feePercent := new(big.Rat).SetFloat64(fee / 100)
	feeValue := new(big.Rat).Mul(wei, feePercent)
	wei.Sub(wei, feeValue)
	wei.Mul(wei, new(big.Rat).SetInt64(shareDiff))
	wei.Quo(wei, new(big.Rat).SetInt64(netDiff))
	shannon := new(big.Rat).SetInt(Shannon)
	inShannon := new(big.Rat).Quo(wei, shannon)
	ppsRate, _ := inShannon.Float64()
	
	return ppsRate
}

func GetShareReward(shareDiff, actualDiff, netDiff int64, potA, potCap, fee float64) float64 {
	// Naive implementation of Pay on Target aka High Variance PPS
	// Reward = Prefix * Factor * PPSRate
	//
	// Prefix = (1-a)/(1-a*wd^(1-a)*X^(a-1))
	// Factor = (min(X,sd)/wd)^a
	// PPSRate is a standard PPS rate at given difficulty
	// wd is always reduced to 1.0 for simplicity
	
	// Reduced values of PoT cap and actual share difficulty
	x, sd := float64(0), float64(0)
	{
		nominalDiff := new(big.Rat).SetInt64(shareDiff)
		SD := new(big.Rat).Quo(new(big.Rat).SetInt64(actualDiff), nominalDiff)
		X  := new(big.Rat).Quo(new(big.Rat).SetInt64(netDiff), nominalDiff)
		X.Mul(X, new(big.Rat).SetFloat64(potCap))
		sd, _ = SD.Float64()
		x, _ = X.Float64()
	}
	
	prefix := (1 - potA) / (1 - potA * refmath.Pow(x, potA - 1))
	factor := refmath.Pow(refmath.Min(x, sd), potA)
	
	// Final calculation
	return prefix * factor * GetPPSRate(shareDiff, netDiff, fee)
}


func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func MustParseDuration(s string) time.Duration {
	value, err := time.ParseDuration(s)
	if err != nil {
		panic("util: Can't parse duration `" + s + "`: " + err.Error())
	}
	return value
}

func String2Big(num string) *big.Int {
	n := new(big.Int)
	n.SetString(num, 0)
	return n
}
