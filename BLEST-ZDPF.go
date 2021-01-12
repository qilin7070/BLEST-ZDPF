func (sch *scheduler) selectBlestPaths(s *session, hasRetransmission bool, hasStreamRetransmission bool, fromPth *path) *path {
	if len(s.paths) <= 1 {
		if !hasRetransmission && !s.paths[protocol.InitialPathID].SendingAllowed() {
			return nil
		}
		return s.paths[protocol.InitialPathID]
	}

	// FIXME Only works at the beginning... Cope with new paths during the connection
	if hasRetransmission && hasStreamRetransmission && fromPth.rttStats.SmoothedRTT() == 0 {
		// Is there any other path with a lower number of packet sent?
		currentQuota := sch.quotas[fromPth.pathID]
		for pathID, pth := range s.paths {
			if pathID == protocol.InitialPathID || pathID == fromPth.pathID {
				continue
			}
			// The congestion window was checked when duplicating the packet
			if sch.quotas[pathID] < currentQuota {
				return pth
			}
		}
	}
	bestPath :=  sch.selectPathLowLatencyf(s, hasRetransmission, hasStreamRetransmission, fromPth)
	if bestPath == nil {
		return bestPath
	}else if bestPath!=nil && bestPath.rttStats.SmoothedRTT()==0{
		return bestPath
	}
	var nlPath *path
	minPath :=  sch.selectPathLowLatencys(s, hasRetransmission, hasStreamRetransmission, fromPth)
	if minPath.rttStats.SmoothedRTT() == 0 {
		return minPath
	}
	if  minPath!=nil && minPath!=bestPath{
		var wait bool
		sch.updateLambda()
		var srtts float64 = sch.srttsCalculation(s,bestPath)
		X := sch.blestschedEstimateBytes(minPath , srtts,s)
		var swnd protocol.ByteCount
		getQueueSize := func(s *stream) (bool,error) {
			if s != nil{
				swnd,_ = s.flowControlManager.SendWindowSize(s.StreamID())
			}
			return true , nil
		}
		s.streamsMap.Iterate(getQueueSize)
		wait = X > float64(swnd)-float64(bestPath.sentPacketHandler.GetBytesInFlight())
		if wait {
			return nlPath
		}
	}
	return bestPath
}