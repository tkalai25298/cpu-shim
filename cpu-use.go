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

const name_space string = "ngp.Consumption"

type Usage struct{
    Time time.Time 	 `json:"time"`
	MeterMPAN string `json:"meter_mpan,omitempty"`
    MAC string		 `json:"macID"`
    DeviceTimestamp string	 `json:"deviceTimestamp"`
    Consumption []Consumption	 `json:"consumption"`
}

type Consumption struct{
    PhaseID uint8	`json:"phaseID"`
    KWh float64		`json:"kwh"`
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
	case "AddCPU":
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
	if len(args) != 2 {
		shim.Error("Incorrect number or arguments")
	}

	name := args[0]
	mpan := args[1]

	key, err:= stub.CreateCompositeKey(name_space,[]string{name,mpan})

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
		Time:            time.Time{},
		MAC:             "",
		DeviceTimestamp: "",
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
	if len(args) != 2 {
		shim.Error("Incorrect number or arguments")
	}

	name := args[0]
	mpan := args[1]

	key, err:= stub.CreateCompositeKey(name_space,[]string{name,mpan})

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
	if len(args) != 7 {
		shim.Error("Incorrect number or arguments")
	}

	name := args[0]
	mpan := args[1]

	key, err:= stub.CreateCompositeKey(name_space,[]string{name,mpan})

	if err != nil {
		return shim.Error(err.Error())
	}

	consumption := []Consumption{}	

	for _,consumedData := range args[4:7] {
		cons := Consumption{}
		err = json.Unmarshal([]byte(consumedData),&cons)
		if err != nil {
			return shim.Error(err.Error())
		}
		consumption = append(consumption,cons)
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
	usageVal.MAC = args[2]

	usageVal.DeviceTimestamp = args[3]

	usageVal.Consumption = consumption
	
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
	if len(args) != 2 {
		shim.Error("Incorrect number or arguments")
	}

	name := args[0]
	mpan := args[1]

	key, err:= stub.CreateCompositeKey(name_space,[]string{name,mpan})

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

		if arrayWritten {
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