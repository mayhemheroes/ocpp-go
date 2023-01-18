package fuzzOcpp

import "strconv"
import "github.com/lorenzodonini/ocpp-go/ocpp2.0.1/types"
import "github.com/lorenzodonini/ocpp-go/ocppj"


func mayhemit(bytes []byte) int {

    var num int
    if len(bytes) > 1 {
        num, _ = strconv.Atoi(string(bytes[0]))

        switch num {
    
        case 0:
            test := types.Now()
            test.UnmarshalJSON(bytes)
            return 0

        case 1:
            content := string(bytes)
            ocppj.ParseJsonMessage(content)
            return 0

        case 2:
            var test ocppj.Endpoint
            content := string(bytes)
            test.GetProfile(content)
            return 0

        case 3:
            var test ocppj.Endpoint
            content := string(bytes)
            test.GetProfileForFeature(content)
            return 0

        default:
            ocppj.ParseRawJsonMessage(bytes)
            return 0

        }
    }
    return 0
}

func Fuzz(data []byte) int {
    _ = mayhemit(data)
    return 0
}