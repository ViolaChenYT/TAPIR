package libstore

import (
	"errors"
	"fmt"
	"net/rpc"
	"sort"
	"sync"
	"time"

	"github.com/ViolaChenYT/TAPIR/common/librpc"
	"github.com/ViolaChenYT/TAPIR/common/storagerpc"
)

type CacheElement struct {
	expires time.Time
	value   string
	vallist []string
}

type VidNode struct {
	HostPort  string
	VirtualID uint32
}

type SortByID []VidNode

func (s SortByID) Len() int {
	return len(s)
}

func (s SortByID) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortByID) Less(i, j int) bool {
	return s[i].VirtualID < s[j].VirtualID
}

type ServerInfoElem struct {
	node storagerpc.Node
	cli  *rpc.Client
}

type libstore struct {
	client     *rpc.Client
	myHostPort string
	mode       LeaseMode
	/* LeaseMode is a debugging flag that determines how the Libstore should request/handle leases
	Never=0, Normal=1, Always=2 */

	storageNodes  []storagerpc.Node
	clientConns   map[string]*rpc.Client
	virtualIDList []VidNode
	queryCts      map[string]int
	cache         map[string]*CacheElement

	queryMux *sync.Mutex
	cacheMux *sync.Mutex
}

// NewLibstore creates a new instance of a TribServer's libstore. masterServerHostPort
// is the master storage server's host:port. myHostPort is this Libstore's host:port
// (i.e. the callback address that the storage servers should use to send back
// notifications when leases are revoked).
//
// The mode argument is a debugging flag that determines how the Libstore should
// request/handle leases. If mode is Never, then the Libstore should never request
// leases from the storage server (i.e. the GetArgs.WantLease field should always
// be set to false). If mode is Always, then the Libstore should always request
// leases from the storage server (i.e. the GetArgs.WantLease field should always
// be set to true). If mode is Normal, then the Libstore should make its own
// decisions on whether or not a lease should be requested from the storage server,
// based on the requirements specified in the project PDF handout.  Note that the
// value of the mode flag may also determine whether or not the Libstore should
// register to receive RPCs from the storage servers.
//
// To register the Libstore to receive RPCs from the storage servers, the following
// line of code should suffice:
//
//	rpc.RegisterName("LeaseCallbacks", librpc.Wrap(libstore))
//
// Note that unlike in the NewTribServer and NewStorageServer functions, there is no
// need to create a brand new HTTP handler to serve the requests (the Libstore may
// simply reuse the TribServer's HTTP handler since the two run in the same process).
func NewLibstore(masterServerHostPort, myHostPort string, mode LeaseMode) (Libstore, error) {

	// join master storage server
	cli, err := rpc.DialHTTP("tcp", masterServerHostPort)
	if err != nil {
		return nil, err
	}

	/*
		Upon creation, an instance of the Libstore will first contact the master storage node using GetServers RPC
		GetServers retrieves a list of available storage servers in the consistent hashing ring
		If GetServers replies with status "NotReady" then not all of the storage servers have joined the ring yet
		If this occurs, your client should sleep for 1 second and retry for up to 5 times (6 total tries)
	*/
	args := &storagerpc.GetServersArgs{}
	reply := &storagerpc.GetServersReply{}
	ready := false

	for attempt := 0; attempt < 6; attempt++ {
		cli.Call("StorageServer.GetServers", args, reply)
		if reply.Status == storagerpc.OK {
			ready = true
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}

	if !ready {
		return nil, errors.New("Hashing ring not complete")
	}

	/*
		When (and if) GetServers replies with status OK, the Libstore will begin to communicate with the storage servers
		via RPC: your Libstore should cache any connections made to the storage servers to ensure efficient communic.
		IE: after opening a connection to a storage server, reuse the connection for subsequent requests
	*/

	ls := &libstore{
		mode:       mode,
		client:     cli,
		myHostPort: myHostPort,
		cache:      make(map[string]*CacheElement),
		cacheMux:   new(sync.Mutex),
		queryMux:   new(sync.Mutex),
	}

	if mode == Normal {
		ls.queryCts = make(map[string]int)
	}

	err = rpc.RegisterName("LeaseCallbacks", librpc.Wrap(ls))
	if err != nil {
		return nil, err
	}

	ls.storageNodes = reply.Servers
	ls.clientConns = make(map[string]*rpc.Client)
	var virtualIDList []VidNode

	for _, server := range reply.Servers {
		var cli *rpc.Client
		cli, err := rpc.DialHTTP("tcp", server.HostPort)
		if err != nil {
			return nil, err
		}
		ls.clientConns[server.HostPort] = cli
		for _, vid := range server.VirtualIDs {
			virtualIDList = append(virtualIDList, VidNode{server.HostPort, vid})
		}
	}

	sort.Sort(SortByID(virtualIDList))
	ls.virtualIDList = virtualIDList

	go ls.CleanCaches()
	return ls, nil
}

func (ls *libstore) CleanCaches() {

	// Every storagerpc.QueryCacheSeconds, go thru caches and delete expired lease's elems

	for {
		timeNow := time.Now()
		ls.cacheMux.Lock()
		for k, v := range ls.cache {
			if v.expires.Sub(timeNow) < 0 {
				delete(ls.cache, k)
			}
		}
		ls.cacheMux.Unlock()

		time.Sleep(storagerpc.QueryCacheSeconds * time.Millisecond * 1000)
	}
}

func (ls *libstore) CheckCaches(key string) *CacheElement {

	// Looks for an element in the caches
	ls.cacheMux.Lock()
	defer ls.cacheMux.Unlock()

	val, ok := ls.cache[key]

	if ok {
		timeNow := time.Now()
		if val.expires.Sub(timeNow) < 0 {
			delete(ls.cache, key)
			return nil
		} else {
			return val
		}
	}
	return nil //cache miss
}

// return the index of the hostport the request should be routed to
func findNextHigher(virtualIDList []VidNode, goal uint32) (int, error) {
	n := len(virtualIDList)

	if goal > virtualIDList[n-1].VirtualID {
		return 0, nil
	}

	for i := 0; i < n; i++ {
		if goal <= virtualIDList[i].VirtualID {
			return i, nil
		}
	}
	return -1, errors.New("problem in finding next higher")

}

func (ls *libstore) Get(key string) (string, error) {
	c := ls.CheckCaches(key)
	if c != nil {
		value := c.value
		return value, nil
	}

	hashedKey := StoreHash(key)
	nextIdx, _ := findNextHigher(ls.virtualIDList, hashedKey)
	nextHostPort := ls.virtualIDList[nextIdx].HostPort
	cli := ls.clientConns[nextHostPort]

	toCache := false
	var args *storagerpc.GetArgs
	var reply storagerpc.GetReply
	// handle lease modes
	if ls.mode == Always {
		args = &storagerpc.GetArgs{Key: key, WantLease: true, HostPort: ls.myHostPort}
		toCache = true
	} else if ls.mode == Never {
		args = &storagerpc.GetArgs{Key: key, WantLease: false, HostPort: ls.myHostPort}
	} else if ls.mode == Normal {
		ls.queryMux.Lock()
		ls.queryCts[key]++
		cts := ls.queryCts[key]
		if cts >= storagerpc.QueryCacheThresh {
			args = &storagerpc.GetArgs{Key: key, WantLease: true, HostPort: ls.myHostPort}
			toCache = true
		} else {
			args = &storagerpc.GetArgs{Key: key, WantLease: false, HostPort: ls.myHostPort}
		}
		ls.queryMux.Unlock()
	}

	// if not in cache

	err := cli.Call("StorageServer.Get", args, &reply)
	for err != nil { // go to next server in ring
		if len(ls.virtualIDList) == 1 {
			return "", err
		}
		nextIdx = (nextIdx + 1) % len(ls.virtualIDList)
		nextHostPort = ls.virtualIDList[nextIdx].HostPort
		cli := ls.clientConns[nextHostPort]
		args.HostPort = nextHostPort
		err = cli.Call("StorageServer.Get", args, &reply)
	}
	if reply.Status != storagerpc.OK {
		return "", errors.New(fmt.Sprintf("%v", reply.Status))
	}

	if toCache {
		ls.cacheMux.Lock()
		ls.cache[key] = &CacheElement{
			expires: time.Now().Add(time.Millisecond * 1000 * time.Duration(reply.Lease.ValidSeconds)),
			value:   reply.Value,
		}
		ls.cacheMux.Unlock()
	}

	return reply.Value, nil
}

func (ls *libstore) Put(key, value string) error {
	hashedKey := StoreHash(key)
	nextIdx, _ := findNextHigher(ls.virtualIDList, hashedKey)
	nextHostPort := ls.virtualIDList[nextIdx].HostPort
	cli := ls.clientConns[nextHostPort]

	args := &storagerpc.PutArgs{Key: key, Value: value}
	var reply storagerpc.PutReply
	err := cli.Call("StorageServer.Put", args, &reply)
	for err != nil { // go to next server in hash ring
		if len(ls.virtualIDList) == 1 {
			return err
		}
		nextIdx = (nextIdx + 1) % len(ls.virtualIDList)
		nextHostPort = ls.virtualIDList[nextIdx].HostPort
		cli := ls.clientConns[nextHostPort]
		err = cli.Call("StorageServer.Put", args, &reply)
	}
	if reply.Status == storagerpc.OK {
		return nil
	}

	return errors.New("Key not found")
}

func (ls *libstore) Delete(key string) error {
	hashedKey := StoreHash(key)
	nextIdx, _ := findNextHigher(ls.virtualIDList, hashedKey)
	nextHostPort := ls.virtualIDList[nextIdx].HostPort
	cli := ls.clientConns[nextHostPort]

	args := &storagerpc.DeleteArgs{Key: key}
	var reply storagerpc.DeleteReply
	err := cli.Call("StorageServer.Delete", args, &reply)
	for err != nil { // go to next server in ring
		if len(ls.virtualIDList) == 1 {
			return err
		}
		nextIdx = (nextIdx + 1) % len(ls.virtualIDList)
		nextHostPort = ls.virtualIDList[nextIdx].HostPort
		cli := ls.clientConns[nextHostPort]
		err = cli.Call("StorageServer.Delete", args, &reply)
	}
	if reply.Status == storagerpc.OK {
		return nil
	}

	return errors.New("Key not found")
}

func (ls *libstore) GetList(key string) ([]string, error) {
	// figure out locking situation -- cant double lock

	if ls.mode != Never {
		//ls.cacheMux.Lock()

		c := ls.CheckCaches(key)
		//ls.cacheMux.Unlock()
		if c != nil {
			value := c.vallist
			return value, nil
		}
	}

	hashedKey := StoreHash(key)
	nextIdx, _ := findNextHigher(ls.virtualIDList, hashedKey)
	nextHostPort := ls.virtualIDList[nextIdx].HostPort
	cli := ls.clientConns[nextHostPort]

	toCache := false
	var args *storagerpc.GetArgs
	var reply storagerpc.GetListReply

	if ls.mode == Always {
		args = &storagerpc.GetArgs{Key: key, WantLease: true, HostPort: ls.myHostPort}
		toCache = true
	} else if ls.mode == Never {
		args = &storagerpc.GetArgs{Key: key, WantLease: false, HostPort: ls.myHostPort}
	} else if ls.mode == Normal {
		ls.queryMux.Lock()
		ls.queryCts[key]++
		cts := ls.queryCts[key]
		if cts >= storagerpc.QueryCacheThresh {
			args = &storagerpc.GetArgs{Key: key, WantLease: true, HostPort: ls.myHostPort}
			toCache = true
		} else {
			args = &storagerpc.GetArgs{Key: key, WantLease: false, HostPort: ls.myHostPort}
		}
		ls.queryMux.Unlock()
	}

	// if not in cache

	err := cli.Call("StorageServer.GetList", args, &reply)
	for err != nil {
		if len(ls.virtualIDList) == 1 {
			return nil, err
		}
		nextIdx = (nextIdx + 1) % len(ls.virtualIDList)
		nextHostPort = ls.virtualIDList[nextIdx].HostPort
		cli := ls.clientConns[nextHostPort]
		args.HostPort = nextHostPort
		err = cli.Call("StorageServer.GetList", args, &reply)
	}
	if reply.Status != storagerpc.OK {
		return nil, errors.New(fmt.Sprintf("%v", reply.Status))
	}
	ret := make([]string, len(reply.Value))
	copy(ret, reply.Value)

	if toCache {
		ls.cacheMux.Lock()
		ls.cache[key] = &CacheElement{
			expires: time.Now().Add(time.Millisecond * 1000 * time.Duration(reply.Lease.ValidSeconds)),
			vallist: ret,
		}
		ls.cacheMux.Unlock()
	}

	return ret, nil
}

func (ls *libstore) RemoveFromList(key, removeItem string) error {
	hashedKey := StoreHash(key)
	nextIdx, _ := findNextHigher(ls.virtualIDList, hashedKey)
	nextHostPort := ls.virtualIDList[nextIdx].HostPort
	cli := ls.clientConns[nextHostPort]

	args := &storagerpc.PutArgs{Key: key, Value: removeItem}
	var reply storagerpc.PutReply
	err := cli.Call("StorageServer.RemoveFromList", args, &reply)
	for err != nil {
		if len(ls.virtualIDList) == 1 {
			return err
		}
		nextIdx = (nextIdx + 1) % len(ls.virtualIDList)
		nextHostPort = ls.virtualIDList[nextIdx].HostPort
		cli := ls.clientConns[nextHostPort]
		err = cli.Call("StorageServer.RemoveFromList", args, &reply)
	}
	if reply.Status == storagerpc.OK {
		return nil
	}

	return errors.New("Key not found (or other error???)")
}

func (ls *libstore) AppendToList(key, newItem string) error {
	hashedKey := StoreHash(key)
	nextIdx, _ := findNextHigher(ls.virtualIDList, hashedKey)
	nextHostPort := ls.virtualIDList[nextIdx].HostPort
	cli := ls.clientConns[nextHostPort]

	args := &storagerpc.PutArgs{Key: key, Value: newItem}
	var reply storagerpc.PutReply
	err := cli.Call("StorageServer.AppendToList", args, &reply)
	for err != nil {
		if len(ls.virtualIDList) == 1 {
			return err
		}
		nextIdx = (nextIdx + 1) % len(ls.virtualIDList)
		nextHostPort = ls.virtualIDList[nextIdx].HostPort
		cli := ls.clientConns[nextHostPort]
		err = cli.Call("StorageServer.AppendToList", args, &reply)
	}
	if reply.Status == storagerpc.OK {
		return nil
	}

	return errors.New("Key not found (or other error???)")
}

/*
RevokeLease is a callback RPC method that is invoked by storage servers when a lease is revoked
Reply with status OK if the key was successfully revoked
Reply with status KeyNotFOund if the key did not exist in the cache
*/
func (ls *libstore) RevokeLease(args *storagerpc.RevokeLeaseArgs, reply *storagerpc.RevokeLeaseReply) error {
	ls.cacheMux.Lock()
	defer ls.cacheMux.Unlock()

	_, ok := ls.cache[args.Key]

	if ok {
		delete(ls.cache, args.Key)
		reply.Status = storagerpc.OK
		return nil
	}

	reply.Status = storagerpc.KeyNotFound
	return nil //cache miss

}
