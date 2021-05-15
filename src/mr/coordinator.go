package mr

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"6.824/asset"
)

// 用于http server template数据传递
type Visual_Data struct {
	Data1 Tree_0
	Data2 Tree_0
	Data3 []Sankey_name
	Data4 []Sankey_link
}

// 三层级联的map任务可视化
// ips -> files
type Tree_2 struct {
	Name      string            `json:"name"`
	ItemStyle map[string]string `json:"itemStyle"`
	Lable     map[string]string `json:"label"`
	State     int               `json:"state"`
	Partition int               `json:"partition"` // only used in reduce work
}

// coordinator -> ips
type Tree_1 struct {
	Name      string            `json:"name"`
	ItemStyle map[string]string `json:"itemStyle"`
	Lable     map[string]string `json:"label"`
	State     int               `json:"state"`
	Children  []Tree_2          `json:"children"`
}

// coordinator
type Tree_0 struct {
	Name     string   `json:"name"`
	Children []Tree_1 `json:"children"`
}

//桑基图 data 结构
type Sankey_names struct {
	Names []Sankey_name
}

// 桑吉图 data 内部的map结构
type Sankey_name struct {
	Name string `json:"name"`
}

// 桑吉图 links 结构
type Sankey_links struct {
	Links []Sankey_link
}

// 桑吉图 link 结构
type Sankey_link struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  int    `json:"value"`
}
type mapwork struct {
	workerid string //id of worker
	filename string //input file name
	state    int    //0 for idle, 1 for in-progress, 2 for completed
}

type reducework struct {
	workerid    string   //id of worker
	state       int      //0 for idle, 1 for in-progress, 2 for completed
	filenames   []string //location of intermediate files
	finishednum int      //number of generated(finished) intermediate files when the reduce worker's last requst came

	ips []string //ip of remote servers. used together with filenames []string
}

type maplist struct {
	list []mapwork
	lock sync.Mutex
}

type reducelist struct {
	list []reducework
	lock sync.Mutex
}

//Coordinator struct
type Coordinator struct {
	// Your definitions here
	nMap         int
	nReduce      int
	maplist      []mapwork
	maplock      sync.Mutex //lock for maplist
	reducelist   []reducework
	reducelock   sync.Mutex //lock for reducelist
	finished     int        //used by Done()
	finishedlock sync.Mutex //lock for finished
}

// Your code here -- RPC handlers for the worker to call.

func (c *Coordinator) backgroundTimer(i int, wtype int) {
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	<-timer.C
	switch wtype {
	case 0: //map work
		c.maplock.Lock()
		if c.maplist[i].state != 2 { //not completed
			//set idle
			c.maplist[i].state = 0
			c.maplist[i].workerid = ""
		}
		c.maplock.Unlock()
	case 1: //reduce work
		c.reducelock.Lock()
		if c.reducelist[i].state != 2 { //not completed
			//set idle
			c.reducelist[i].state = 0
			c.reducelist[i].workerid = ""
			c.reducelist[i].finishednum = 0
		}
		c.reducelock.Unlock()
	}
}

//handler for AskWork
//assign work for workers
//if no work to assign, set reply.wtype to 2
//this will inform the worker to exit
func (c *Coordinator) WorkHandler(args *RequestArgs, reply *WorkRequestReply) error {
	if args.Rtype != 0 {
		return fmt.Errorf("rtype wrong")
	}

	//first check maplist in-progress
	c.maplock.Lock()
	for i := range c.maplist {
		if c.maplist[i].workerid == args.Workerid && c.maplist[i].state == 1 { //in-progress

			c.maplist[i].state = 2 //deem the request as a report for finish
			//add files to reducelist!
			filename := c.maplist[i].filename
			c.maplock.Unlock()

			c.reducelock.Lock()
			for i := 0; i < c.nReduce; i++ {
				ifilename := fmt.Sprintf("mr-%s-%d", filepath.Base(filename), i)
				c.reducelist[i].filenames = append(c.reducelist[i].filenames, ifilename)
				c.reducelist[i].ips = append(c.reducelist[i].ips, args.Workerid)
			}
			c.reducelock.Unlock()
			goto checkidle
		}
	}
	c.maplock.Unlock()

	//then check reducelist in-progress
	c.reducelock.Lock()
	for i := range c.reducelist {
		if c.reducelist[i].workerid == args.Workerid && c.reducelist[i].state == 1 { //in-progress
			c.reducelist[i].state = 2 //deem the request as a report for finish
			c.finishedlock.Lock()
			c.finished++
			c.finishedlock.Unlock()
			break
		}
	}
	c.reducelock.Unlock()

checkidle:
	//first check maplist idle
	c.maplock.Lock()
	flag := false

	for i := range c.maplist {
		if c.maplist[i].state == 0 { //idle
			go c.backgroundTimer(i, 0) //set timer

			c.maplist[i].state = 1
			c.maplist[i].workerid = args.Workerid

			reply.Wtype = 0                        //map work type
			reply.NMap = c.nMap                    //number of map tasks
			reply.NReduce = c.nReduce              //number of reduce tasks
			reply.Filename = c.maplist[i].filename //map work's input filename
			c.maplock.Unlock()
			return nil
		}
		if c.maplist[i].state == 1 { //in-progress
			flag = true
		}
	}
	c.maplock.Unlock()

	//then check reducelist idle
	c.reducelock.Lock()
	for i := range c.reducelist {
		if c.reducelist[i].state == 0 { //idle
			go c.backgroundTimer(i, 1) //set timer

			c.reducelist[i].state = 1
			c.reducelist[i].workerid = args.Workerid

			reply.Wtype = 1           //reduce work type
			reply.NMap = c.nMap       //number of map tasks
			reply.NReduce = c.nReduce //number of reduce tasks
			reply.Rnumber = i         //reduce partition number
			c.reducelock.Unlock()
			return nil
		}
		if c.reducelist[i].state == 1 { //in-progress
			flag = true
		}
	}
	c.reducelock.Unlock()

	//do not need more workers
	if flag { //can not exit because there is still in-progress workers.wait for a while
		reply.Wtype = 3
	} else { //no workers in progress, this worker can exit
		reply.Wtype = 2
	}
	return nil

}

//handler for AskFile
//send intermediate file names to reduce worker
func (c *Coordinator) FileHandler(args *RequestArgs, reply *FileRequestReply) error {
	if args.Rtype != 1 {
		return fmt.Errorf("rtype wrong")
	}
	c.reducelock.Lock()
	for i := range c.reducelist {
		if c.reducelist[i].workerid == args.Workerid && c.reducelist[i].state == 1 { //in-progress
			reply.Status = 0 //normal
			reply.Filenames = c.reducelist[i].filenames[c.reducelist[i].finishednum:]
			reply.Ips = c.reducelist[i].ips[c.reducelist[i].finishednum:]
			c.reducelist[i].finishednum = len(c.reducelist[i].filenames)
			c.reducelock.Unlock()
			return nil
		}
	}
	//workerid := args.Workerid
	c.reducelock.Unlock()

	//if code goes here
	//it may because the reducer has been deemed dead, but it actually not
	//we need to inform it the exception,
	//let it discard currunt work, and ask for a new one
	reply.Status = 1 //exception
	reply.Filenames = []string{}
	//for output simplicity, the following Printf sentence is commented out.
	//fmt.Printf("workerid %d not found in in-progress workers. This message can be ignored\n", workerid)
	return nil
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	//sockname := coordinatorSock()
	//os.Remove(sockname)
	//l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func (c *Coordinator) Get_visual_data() Visual_Data {

	c.maplock.Lock()
	map_list := make([]mapwork, len(c.maplist))
	copy(map_list, c.maplist)
	c.maplock.Unlock()

	c.reducelock.Lock()
	reduce_list := make([]reducework, len(c.reducelist))
	copy(reduce_list, c.reducelist)
	c.reducelock.Unlock()

	map_subtree := make(map[string][]Tree_2)
	map_tree := Tree_0{}

	for i := range map_list {
		map_work := map_list[i]
		ip := "Mapper " + map_work.workerid
		filename := map_work.filename
		state := map_work.state
		if state == 0 {
			continue
		} else if state == 1 {
			map_subtree[ip] = append(map_subtree[ip], Tree_2{Name: filename, ItemStyle: asset.Itemstyle_setting_notdone, State: 0})
		} else {
			map_subtree[ip] = append(map_subtree[ip], Tree_2{Name: filename, ItemStyle: asset.Itemstyle_setting_done, State: 1})
		}
	}

	// ip排序
	var ips []string
	for ip := range map_subtree {
		ips = append(ips, ip)
	}
	sort.Strings(ips)

	map_tree.Name = "MapWork"
	for _, ip := range ips {
		files := map_subtree[ip]
		sub_tree := Tree_1{}
		sub_tree.Name = ip
		sub_tree.Children = files
		map_tree.Children = append(map_tree.Children, sub_tree)
	}

	// 对reduce work重复操作
	reduce_subtree := make(map[string][]Tree_2)
	reduce_tree := Tree_0{}
	for i := range reduce_list { //initialze map works' input file names
		reduce_work := reduce_list[i]
		ip := "Reducer " + reduce_work.workerid
		state := reduce_work.state
		filenames := reduce_work.filenames
		files_done := reduce_work.finishednum
		if state == 0 {
			continue
		} else if state == 1 {
			for j := range filenames {
				filename := filenames[j]
				if j < files_done {
					reduce_subtree[ip] = append(reduce_subtree[ip], Tree_2{Name: filename, ItemStyle: asset.Itemstyle_setting_done, State: 1, Partition: i})
				} else {
					reduce_subtree[ip] = append(reduce_subtree[ip], Tree_2{Name: filename, ItemStyle: asset.Itemstyle_setting_notdone, State: 0, Partition: i})
				}
			}
		} else {
			for j := range filenames {
				filename := filenames[j]
				reduce_subtree[ip] = append(reduce_subtree[ip], Tree_2{Name: filename, ItemStyle: asset.Itemstyle_setting_done, State: 1, Partition: i})
			}
		}
	}

	// ip排序
	ips = []string{}
	for ip := range reduce_subtree {
		ips = append(ips, ip)
	}
	sort.Strings(ips)

	reduce_tree.Name = "ReduceWork"
	for _, ip := range ips {
		files := reduce_subtree[ip]
		sub_tree := Tree_1{}
		sub_tree.Name = ip
		sub_tree.Children = files
		reduce_tree.Children = append(reduce_tree.Children, sub_tree)
	}

	// 桑基图
	sankey_names := Sankey_names{}
	sankey_links := Sankey_links{}
	n_reduce := c.nReduce

	for _, ip := range map_tree.Children {
		sankey_names.Names = append(sankey_names.Names, Sankey_name{Name: ip.Name})
		for _, inputfile := range ip.Children {
			sankey_names.Names = append(sankey_names.Names, Sankey_name{Name: inputfile.Name})
			sankey_links.Links = append(sankey_links.Links, Sankey_link{Source: inputfile.Name, Target: ip.Name, Value: 1})
			for i := 0; i < n_reduce; i++ {
				intermediatefile := fmt.Sprintf("mr-%s-%d", inputfile.Name, i)
				sankey_links.Links = append(sankey_links.Links, Sankey_link{Source: ip.Name, Target: intermediatefile, Value: 1})
			}
		}
	}
	var partition_prev = -1
	for _, ip := range reduce_tree.Children {
		sankey_names.Names = append(sankey_names.Names, Sankey_name{Name: ip.Name})
		for _, intermediatefile := range ip.Children {
			sankey_names.Names = append(sankey_names.Names, Sankey_name{Name: intermediatefile.Name})
			sankey_links.Links = append(sankey_links.Links, Sankey_link{Source: intermediatefile.Name, Target: ip.Name, Value: 1})
			if intermediatefile.Partition != partition_prev {
				oname := fmt.Sprintf("mr-out-%d", intermediatefile.Partition)
				sankey_names.Names = append(sankey_names.Names, Sankey_name{Name: oname})
				sankey_links.Links = append(sankey_links.Links, Sankey_link{Source: ip.Name, Target: oname, Value: 1})
			}
			partition_prev = intermediatefile.Partition
		}
	}

	all_data := Visual_Data{Data1: map_tree, Data2: reduce_tree, Data3: sankey_names.Names, Data4: sankey_links.Links}

	return all_data

}

func (c *Coordinator) start_web() {
	http.HandleFunc("/data", c.dataserver)
	http.HandleFunc("/", c.httpserver)
	http.ListenAndServe(":8081", nil)
}

// YifanLu here
func (c *Coordinator) httpserver(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //解析参数，默认是不会解析的
	t, _ := template.ParseFiles("dashboard.html")
	t.Execute(w, nil)
}

func (c *Coordinator) dataserver(w http.ResponseWriter, r *http.Request) {
	visual_Data := c.Get_visual_data()
	json_obj, _ := json.Marshal(visual_Data)
	w.Header().Set("content-type", "application/json")
	w.Write(json_obj)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {

	// Your code here.
	c.finishedlock.Lock()
	done := c.finished == c.nReduce
	c.finishedlock.Unlock()

	return done
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	c.nMap = len(files)
	c.nReduce = nReduce

	c.maplist = make([]mapwork, c.nMap)
	c.reducelist = make([]reducework, c.nReduce)

	for i := range c.maplist { //initialze map works' input file names
		c.maplist[i].filename = files[i]
	}

	go c.start_web()
	c.server()
	return &c
}
