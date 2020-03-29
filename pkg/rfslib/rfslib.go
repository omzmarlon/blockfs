package rfslib

import (
	"fmt"
	"net/rpc"
)

// A Record is the unit of file access (reading/appending) in RFS.
type Record [512]byte

////////////////////////////////////////////////////////////////////////////////////////////
// <ERROR DEFINITIONS>

// These type definitions allow the application to explicitly check
// for the kind of error that occurred. Each API call below lists the
// errors that it is allowed to raise.
//
// Also see:
// https://blog.golang.org/error-handling-and-go
// https://blog.golang.org/errors-are-values

// Contains minerAddr
type DisconnectedError string

func (e DisconnectedError) Error() string {
	return fmt.Sprintf("RFS: Disconnected from the miner [%s]", string(e))
}

// Contains filename. The *only* constraint on filenames in RFS is
// that must be at most 64 bytes long.
type BadFilenameError string

func (e BadFilenameError) Error() string {
	return fmt.Sprintf("RFS: Filename [%s] has the wrong length", string(e))
}

// Contains filename
type FileDoesNotExistError string

func (e FileDoesNotExistError) Error() string {
	return fmt.Sprintf("RFS: Cannot open file [%s] in D mode as it does not exist locally", string(e))
}

// Contains filename
type FileExistsError string

func (e FileExistsError) Error() string {
	return fmt.Sprintf("RFS: Cannot create file with filename [%s] as it already exists", string(e))
}

// Contains filename
type FileMaxLenReachedError string

func (e FileMaxLenReachedError) Error() string {
	return fmt.Sprintf("RFS: File [%s] has reached its maximum length", string(e))
}

// </ERROR DEFINITIONS>
////////////////////////////////////////////////////////////////////////////////////////////

// Represents a connection to the RFS system.
type RFS interface {
	// Creates a new empty RFS file with name fname.
	//
	// Can return the following errors:
	// - DisconnectedError
	// - FileExistsError
	// - BadFilenameError
	CreateFile(fname string) (err error)

	// Returns a slice of strings containing filenames of all the
	// existing files in RFS.
	//
	// Can return the following errors:
	// - DisconnectedError
	ListFiles() (fnames []string, err error)

	// Returns the total number of records in a file with filename
	// fname.
	//
	// Can return the following errors:
	// - DisconnectedError
	// - FileDoesNotExistError
	TotalRecs(fname string) (numRecs uint16, err error)

	// Reads a record from file fname at position recordNum into
	// memory pointed to by record. Returns a non-nil error if the
	// read was unsuccessful. If a record at this index does not yet
	// exist, this call must block until the record at this index
	// exists, and then return the record.
	//
	// Can return the following errors:
	// - DisconnectedError
	// - FileDoesNotExistError
	ReadRec(fname string, recordNum uint16, record *Record) (err error)

	// Appends a new record to a file with name fname with the
	// contents pointed to by record. Returns the position of the
	// record that was just appended as recordNum. Returns a non-nil
	// error if the operation was unsuccessful.
	//
	// Can return the following errors:
	// - DisconnectedError
	// - FileDoesNotExistError
	// - FileMaxLenReachedError
	AppendRec(fname string, record *Record) (recordNum uint16, err error)
}

type RFSObj struct {
	minerConnection *rpc.Client
	minerAddr       string
}

type RequestMessage struct {
	OpName    string
	Fname     string
	RecordNum uint16
	// we changed this to Record not *Record because we can't send pointers over the wire
	Record Record
}

type ResponseMessage struct {
	NumRecs   uint16
	Record    Record
	Fnames    []string
	RecordNum uint16
	Error     string
}

func (rfsObj RFSObj) CreateFile(fname string) (err error) {

	if !isValidFileName(fname) {
		return BadFilenameError(fname)
	}

	message := &RequestMessage{"CREATE", fname, 0, Record{}}

	_, err = communicateWithMiner(message, rfsObj, "ReqHandler.HandleRemoteRequest")

	if err != nil {
		return err
	}

	return nil
}

func (rfsObj RFSObj) ListFiles() (fnames []string, err error) {

	message := &RequestMessage{"LIST", "", 0, Record{}}

	res, err := communicateWithMiner(message, rfsObj, "ReqHandler.HandleLocalRequest")

	if err != nil {
		return nil, err
	}

	return res.Fnames, nil

}

func (rfsObj RFSObj) TotalRecs(fname string) (numRecs uint16, err error) {

	message := &RequestMessage{"TOTALREC", fname, 0, Record{}}

	res, err := communicateWithMiner(message, rfsObj, "ReqHandler.HandleLocalRequest")

	if err != nil {
		return 0, err
	}

	return res.NumRecs, nil
}

func (rfsObj RFSObj) ReadRec(fname string, recordNum uint16, record *Record) (err error) {

	message := &RequestMessage{"READREC", fname, recordNum, Record{}}

	res, err := communicateWithMiner(message, rfsObj, "ReqHandler.HandleLocalRequest")

	if err != nil {
		return err
	}

	*record = res.Record

	return nil
}

func (rfsObj RFSObj) AppendRec(fname string, record *Record) (recordNum uint16, err error) {

	if len(*record) > 512 {
		return 0, FileMaxLenReachedError(fname)
	}

	message := &RequestMessage{"APPENDREC", fname, 0, *record}

	res, err := communicateWithMiner(message, rfsObj, "ReqHandler.HandleRemoteRequest")

	if err != nil {
		return 0, err
	}

	return res.RecordNum, nil
}

// The constructor for a new RFS object instance. Takes the miner's
// IP:port address string as parameter, and the localAddr which is the
// local IP:port to use to establish the connection to the miner.
//
// The returned rfs instance is singleton: an application is expected
// to interact with just one rfs at a time.
//
// This call should only succeed if the connection to the miner
// succeeds. This call can return the following errors:
// - Networking errors related to localAddr or minerAddr
func Initialize(localAddr string, minerAddr string) (rfs RFS, err error) {
	//TODO figure out what localAddr is used for
	minerConn, err := resolveAndCreateTCPConnection(minerAddr)

	if err != nil {
		return nil, DisconnectedError(minerAddr)
	}

	rfs = RFSObj{minerConn, minerAddr}

	return rfs, nil
}

func resolveAndCreateTCPConnection(minerAddr string) (*rpc.Client, error) {

	/*lAddr, err := resolveTCPAddr(localAddr)
	if err != nil {
		return nil, errors.New("Can not resolve" + localAddr)
	}

	rAddr, err := resolveTCPAddr(minerAddr)
	if err != nil {
		return nil, errors.New("Can not resolve" + minerAddr)
	}*/

	return rpc.DialHTTP("tcp", minerAddr)

}

/*func resolveTCPAddr(addr string) (*net.TCPAddr, error) {
	return net.ResolveTCPAddr("tcp", addr)
}*/

func isValidFileName(fileName string) bool {
	bytes := []byte(fileName)
	return len(bytes) <= 64
}
