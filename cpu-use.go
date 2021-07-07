package main

import (
	"encoding/json"
	"bytes"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type SimpleChaincode struct{
}

var name_space string = "org.cpu-use.Usage"

type Usage struct{
    Time time.Time 	 `json:"time"`
    MAC string		 `json:"mac"`
    DeviceTimestamp time.Time	 `json:"dts"`
    Consumption []Consumption	 `json:"consumption"`
}

type Consumption struct{
    PhaseID uint8	`json:"phaseID"`
    KWh float32		`json:"kwh"`
}

func main() {

	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Println("Error starting chaincode server")
	}
}

// Init executes at the start
func (c *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Chaincode Initiated")
	return shim.Success(nil)
}

// Invoke acts as a router
func (c *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fun, args := stub.GetFunctionAndParameters()

	fmt.Println("Executing => "+fun)

	switch fun{
	case "init":
		return c.init(stub,args)
	case "AddCpu":
		return c.AddCpu(stub,args)
	case "AddUsage":
		return c.AddUsage(stub,args)
	case "GetUsage":
		return c.GetUsage(stub,args)
	case "GetHistory":
		return c.GetHistory(stub,args)
	default:
		return shim.Error("Not a vaild function")	
	}
}

// init just for instantiate
func (c *SimpleChaincode) init(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	fmt.Println("DONE !!!")
	return shim.Success(nil)
}

//AddCpu register a cpu
func (c *SimpleChaincode) AddCpu(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	if len(args) != 1 {
		shim.Error("Incorrect number or arguments")
	}

	name := args[0]
	key, err:= stub.CreateCompositeKey(name_space,[]string{name})

	if err != nil {
		return shim.Error(err.Error())
	}

	usageGet, err:= stub.GetState(key)

	if err != nil {
		return shim.Error(err.Error())
	} else if usageGet != nil {
		return shim.Error("Asset already exists")
	}

	usageVal := &Usage{
		Time:       time.Time{},
		MAC:             "",
		DeviceTimestamp: time.Time{},
		Consumption:     []Consumption{},
	}

	usageByte, err := json.Marshal(usageVal)
	if err != nil{
		return shim.Error(err.Error())
	}

	err = stub.PutState(key,usageByte)

	if err != nil{
		return shim.Error(err.Error())
	}

	return shim.Success(usageByte)
}

// GetUsage returns stored value
func (c *SimpleChaincode) GetUsage(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		shim.Error("Incorrect number or arguments")
	}

	name := args[0]
	key, err:= stub.CreateCompositeKey(name_space,[]string{name})

	if err != nil {
		return shim.Error(err.Error())
	}

	usageGet, err:= stub.GetState(key)

	if err != nil {
		return shim.Error(err.Error())
	} else if usageGet == nil {
		return shim.Error("Empty asset")
	}

	return shim.Success(usageGet)
}

// AddUsage to update the cpu asset
func (c *SimpleChaincode) AddUsage(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 6 {
		shim.Error("Incorrect number or arguments")
	}

	name := args[0]
	key, err:= stub.CreateCompositeKey(name_space,[]string{name})

	if err != nil {
		return shim.Error(err.Error())
	}

	consumption := []Consumption{}

	consumed,err := json.Marshal([]string{args[3],args[4],args[5]})
	if err != nil {
		fmt.Println("consumption array marshal err")
		return shim.Error(err.Error())
	}
	
	
	err = json.Unmarshal(consumed,&consumption)
	if err != nil {
		fmt.Println("consumption array unmarshal err")
		return shim.Error(err.Error())
	}

	usageGet, err := stub.GetState(key)

	if err != nil {
		return shim.Error(err.Error())
	} else if usageGet == nil {
		return shim.Error("Empty asset")
	}

	var usageVal Usage

	err = json.Unmarshal([]byte(usageGet), &usageVal)

	if err != nil {
		return shim.Error(err.Error())
	}

	usageVal.Time = time.Now()
	usageVal.MAC = args[1]

	dts,err := time.Parse("ANSIC",args[2])
	if err != nil {
		return shim.Error(err.Error())
	}
	usageVal.DeviceTimestamp = dts

	usageVal.Consumption = append(usageVal.Consumption, consumption[0],consumption[1],consumption[2])
	
	usageByte, err := json.Marshal(usageVal)
	if err != nil{
		return shim.Error(err.Error())
	}

	err = stub.PutState(key,usageByte)

	if err != nil{
		return shim.Error(err.Error())
	}

	return shim.Success(usageByte)
}

// GetHistory returns entire history of the asset
func (c *SimpleChaincode) GetHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		shim.Error("Incorrect number or arguments")
	}

	name := args[0]
	key, err:= stub.CreateCompositeKey(name_space,[]string{name})

	if err != nil {
		return shim.Error(err.Error())
	}

	resultsIterator, err := stub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	arrayWritten := false

	for resultsIterator.HasNext(){
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if arrayWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Value\":")
		if response.IsDelete {
			buffer.WriteString("Deleted")
		} else {
			buffer.WriteString(string(response.Value))
		}
		buffer.WriteString("}")

		arrayWritten = true
	}
	buffer.WriteString("]")

	fmt.Println("History is : "+buffer.String())

	return shim.Success(buffer.Bytes())
}