package http

//
//type rpcServer struct {
//	service *Service
//}
//
//// WorkerList contains list of workers.
//type WorkerList struct {
//	// Workers is list of workers.
//	Workers []utils.Worker `json:"workers"`
//}
//
//// Reset resets underlying RR worker pool and restarts all of it's workers.
//func (rpc *rpcServer) Reset(reset bool, r *string) error {
//	if rpc.service.srv == nil {
//		return errors.New("no http server")
//	}
//
//	logrus.Info("http: restarting worker pool")
//	*r = "OK"
//
//	err := rpc.service.srv.rr.Reset()
//	if err != nil {
//		logrus.Errorf("http: %s", err)
//	}
//
//	return err
//}
//
//// Workers returns list of active workers and their stats.
//func (rpc *rpcServer) Workers(list bool, r *WorkerList) error {
//	if rpc.service.srv == nil {
//		return errors.New("no http server")
//	}
//
//	r.Workers = utils.FetchWorkers(rpc.service.srv.rr)
//	return nil
//}
//
//// Worker provides information about specific worker.
//type Worker struct {
//	// Pid contains process id.
//	Pid int `json:"pid"`
//
//	// Status of the worker.
//	Status string `json:"status"`
//
//	// Number of worker executions.
//	NumExecs uint64 `json:"numExecs"`
//
//	// Created is unix nano timestamp of worker creation time.
//	Created int64 `json:"created"`
//
//	// Updated is unix nano timestamp of last worker execution.
//	Updated int64 `json:"updated"`
//}
//
//// FetchWorkers fetches list of workers from RR Server.
//func FetchWorkers(srv *roadrunner.Server) (result []Worker) {
//	for _, w := range srv.Workers() {
//		state := w.State()
//		result = append(result, Worker{
//			Pid:      *w.Pid,
//			Status:   state.String(),
//			NumExecs: state.NumExecs(),
//			Created:  w.Created.UnixNano(),
//			Updated:  state.Updated().UnixNano(),
//		})
//	}
//
//	return
//}
