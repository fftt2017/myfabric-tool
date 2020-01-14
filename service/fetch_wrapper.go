package service

import (
	"encoding/base64"
	"github.com/buger/jsonparser"
	"log"
	"strconv"
)

func WrapFetchResult(fetchResult string) (string) {

	fetchResultArr := []byte(fetchResult)
	resultByte :=[]byte(`
			{
				"data_hash":"",
				"number":"",
				"previous_hash":"",
				"validation_code":"",
				"count":"",
				"datas":[]
			}
	`)
	resultByte = processHeaderAndMeta(resultByte,fetchResultArr)
	resultByte = processDatas(resultByte,fetchResultArr)
	result := string(resultByte)
	log.Println(result)
	return result
}

func processHeaderAndMeta(jsonResult []byte,fetchResult []byte)([]byte){
	data_hash,_:= jsonparser.GetString(fetchResult,"header","data_hash")
	number,_,_,_ := jsonparser.Get(fetchResult,"header","number")
	previous_hash,_ := jsonparser.GetString(fetchResult,"header","previous_hash")

	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+data_hash+`"`),"data_hash")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(number),"number")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+previous_hash+`"`),"previous_hash")

	validation_code_bytes,_,_,_ := jsonparser.Get(fetchResult,"metadata","metadata","[2]")
	codeByteArr := []byte("")
	codeByteArr,_ = base64.StdEncoding.DecodeString(string(validation_code_bytes))
	codeArr := make([]string, len(codeByteArr))
	for i := 0; i < len(codeByteArr); i++ {
		code := int(codeByteArr[i])
		switch code {
		case 0: codeArr[i] = "VALID"
		case 1: codeArr[i] = "NIL_ENVELOPE"
		case 2 : codeArr[i] = "BAD_PAYLOAD"
		case 3 : codeArr[i] = "BAD_COMMON_HEADER"
		case 4 : codeArr[i] = "BAD_CREATOR_SIGNATURE"
		case 5 : codeArr[i] = "INVALID_ENDORSER_TRANSACTION"
		case 6 : codeArr[i] = "INVALID_CONFIG_TRANSACTION"
		case 7 : codeArr[i] = "UNSUPPORTED_TX_PAYLOAD"
		case 8 : codeArr[i] = "BAD_PROPOSAL_TXID"
		case 9 : codeArr[i] = "DUPLICATE_TXID"
		case 10 : codeArr[i] = "ENDORSEMENT_POLICY_FAILURE"
		case 11 : codeArr[i] = "MVCC_READ_CONFLICT"
		case 12 : codeArr[i] = "PHANTOM_READ_CONFLICT"
		case 13 : codeArr[i] = "UNKNOWN_TX_TYPE"
		case 14 : codeArr[i] = "TARGET_CHAIN_NOT_FOUND"
		case 15 : codeArr[i] = "MARSHAL_TX_ERROR"
		case 16 : codeArr[i] = "NIL_TXACTION"
		case 17 : codeArr[i] = "EXPIRED_CHAINCODE"
		case 18 : codeArr[i] = "CHAINCODE_VERSION_CONFLICT"
		case 19 : codeArr[i] = "BAD_HEADER_EXTENSION"
		case 20 : codeArr[i] = "BAD_CHANNEL_HEADER"
		case 21 : codeArr[i] = "BAD_RESPONSE_PAYLOAD"
		case 22 : codeArr[i] = "BAD_RWSET"
		case 23 : codeArr[i] = "ILLEGAL_WRITESET"
		case 24 : codeArr[i] = "INVALID_WRITESET"
		case 254 : codeArr[i] = "NOT_VALIDATED"
		case 255 : codeArr[i] = "INVALID_OTHER_REASON"
		default: codeArr[i] = "unrecognized value " + string(codeByteArr[i])
		}
	}
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+codeArr[0]+`"`),"validation_code")
	return jsonResult
}
func processDatas(jsonResult []byte,fetchResult []byte)([]byte){
	dataResult := ""
	var index int = 0
	jsonparser.ArrayEach(fetchResult, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		log.Println("data is :"+string(value))
		if index == 0 {
			dataResult = string(processTransaction(value))
		}else{
			dataResult = dataResult + "," + string(processTransaction(value))
		}
		index++

	}, "data", "data")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(strconv.Itoa(index)),"count")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte("["+dataResult+"]"),"datas")
	log.Print(string(jsonResult))
	return (jsonResult)
}
func processTransaction(dataJson []byte)([]byte){
	jsonResult :=[]byte(`
			{
				"channel_id":"",
				"timestamp":"",
				"tx_id":"",
				"type":"",
				"creatorId":"",
				"mspid":"",
				"actions":[]
			}
	`)
	channel_id,_:= jsonparser.GetString(dataJson,"payload","header","channel_header","channel_id")
	timestamp,_:= jsonparser.GetString(dataJson,"payload","header","channel_header","timestamp")
	tx_id,_:= jsonparser.GetString(dataJson,"payload","header","channel_header","tx_id")
	type_,_ := jsonparser.GetInt(dataJson,"payload","header","channel_header","type")
	creatorId,_ := jsonparser.GetString(dataJson,"payload","header","signature_header","creator","id_bytes")
	mspid,_ := jsonparser.GetString(dataJson,"payload","header","signature_header","creator","mspid")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+channel_id+`"`),"channel_id")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+timestamp+`"`),"timestamp")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+tx_id+`"`),"tx_id")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+strconv.FormatInt(type_,10)+`"`),"type")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+creatorId+`"`),"creatorId")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+mspid+`"`),"mspid")

	var actions string = ""
	var index int = 0
	jsonparser.ArrayEach(dataJson, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		log.Println("data is :"+string(value))
		if index == 0 {
			actions = actions + string(processAction(value))
		}else{
			actions = actions + "," + string(processAction(value))
		}
		index++

	}, "payload", "data", "actions")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte("["+actions+"]"),"actions")
	return jsonResult
}

func processAction(dataJson []byte)([]byte){
	jsonResult :=[]byte(`
			{
				"endorsements":[],
				"chaincode_proposal":{
					"chaincode_name":"",
					"type":"",
					"input_args":[]
				},
				"proposal_response":{
					"chaincode_name":"",
					"status":"",
					"message":"",
					"ns_rwset":[]
				}
				
			}
	`)

	var endorsements string = ""
	var index int = 0
	jsonparser.ArrayEach(dataJson, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		log.Println(string(value))
		endorser,_ := jsonparser.GetString(value,"endorser")
		log.Println("endorser is :"+string(endorser))
		if index == 0 {
			endorsements = `"` + endorser + `"`
		}else{
			endorsements = endorsements + `,"` + endorser + `"`
		}
		index++

	}, "payload", "action", "endorsements")

	chaincode_proposal_name,_ := jsonparser.GetString(dataJson, "payload", "chaincode_proposal_payload", "input", "chaincode_spec", "chaincode_id", "name")
	chaincode_proposal_type,_ := jsonparser.GetString(dataJson, "payload", "chaincode_proposal_payload", "input", "chaincode_spec", "type")
	input_args,_,_,_:= jsonparser.Get(dataJson, "payload", "chaincode_proposal_payload", "input", "chaincode_spec", "input", "args")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte("["+endorsements+"]"),"endorsements")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+chaincode_proposal_name+`"`),"chaincode_proposal","chaincode_name")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+chaincode_proposal_type+`"`),"chaincode_proposal","type")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(input_args),"chaincode_proposal","input_args")

	response_chaincode_name,_ := jsonparser.GetString(dataJson, "payload", "action", "proposal_response_payload", "extension", "chaincode_id", "name")
	response_status,_:= jsonparser.GetInt(dataJson, "payload", "action", "proposal_response_payload", "extension", "response", "status")
	response_msg,_:= jsonparser.GetString(dataJson, "payload", "action", "proposal_response_payload", "extension", "response", "message")
	ns_rwset,_,_,_:=jsonparser.Get(dataJson, "payload", "action", "proposal_response_payload", "extension", "results", "ns_rwset")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+response_chaincode_name+`"`),"proposal_response","chaincode_name")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+strconv.FormatInt(response_status,10)+`"`),"proposal_response","status")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(`"`+response_msg+`"`),"proposal_response","message")
	jsonResult,_ = jsonparser.Set(jsonResult,[]byte(ns_rwset),"proposal_response","ns_rwset")

	return jsonResult
}

