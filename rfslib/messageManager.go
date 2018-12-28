package rfslib

func communicateWithMiner(message *RequestMessage, rfsObj RFSObj, method string) (ResponseMessage, error) {

	reply := ResponseMessage{}
	//call is synchronous
	err := rfsObj.minerConnection.Call(method, message, &reply)

	if err != nil {
		// error for Call method of rpc
		return ResponseMessage{0, Record{}, nil, 0, ""}, DisconnectedError(rfsObj.minerAddr)
	}

	//other errors are returned here
	return reply, getErrorType(reply.Error, message.Fname, rfsObj.minerAddr)
}

func getErrorType(errMessage string, fname string, minerAddr string) error {

	switch errMessage {
	case "FILEEXISTS":
		return FileExistsError(fname)
	case "FILEDNE":
		return FileDoesNotExistError(fname)
	case "FILEMLR":
		return FileMaxLenReachedError(fname)
	case "DISCONNECTED":
		return DisconnectedError(minerAddr)
	}

	return nil
}
