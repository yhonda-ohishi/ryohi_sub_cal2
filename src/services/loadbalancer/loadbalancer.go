package loadbalancer

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/your-org/ryohi-router/src/models"
)

// LoadBalancer interface for different load balancing algorithms
type LoadBalancer interface {
	Next() *models.EndpointConfig
	MarkHealthy(endpoint *models.EndpointConfig)
	MarkUnhealthy(endpoint *models.EndpointConfig)
}

// New creates a new load balancer based on the algorithm
func New(config *models.LoadBalancerConfig, endpoints []models.EndpointConfig) (LoadBalancer, error) {
	switch config.Algorithm {
	case "round-robin", "":
		return NewRoundRobin(endpoints), nil
	case "weighted":
		return NewWeighted(endpoints), nil
	case "least-conn":
		return NewLeastConnections(endpoints), nil
	case "random":
		return NewRandom(endpoints), nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", config.Algorithm)
	}
}

// RoundRobin implements round-robin load balancing
type RoundRobin struct {
	endpoints []models.EndpointConfig
	current   uint32
	mutex     sync.RWMutex
}

// NewRoundRobin creates a new round-robin load balancer
func NewRoundRobin(endpoints []models.EndpointConfig) *RoundRobin {
	return &RoundRobin{
		endpoints: endpoints,
		current:   0,
	}
}

// Next returns the next endpoint in round-robin fashion
func (rr *RoundRobin) Next() *models.EndpointConfig {
	rr.mutex.RLock()
	defer rr.mutex.RUnlock()

	if len(rr.endpoints) == 0 {
		return nil
	}

	// Find healthy endpoints
	healthyEndpoints := make([]models.EndpointConfig, 0)
	for _, ep := range rr.endpoints {
		if ep.Healthy {
			healthyEndpoints = append(healthyEndpoints, ep)
		}
	}

	if len(healthyEndpoints) == 0 {
		return nil
	}

	// Get next endpoint
	index := atomic.AddUint32(&rr.current, 1) % uint32(len(healthyEndpoints))
	return &healthyEndpoints[index]
}

// MarkHealthy marks an endpoint as healthy
func (rr *RoundRobin) MarkHealthy(endpoint *models.EndpointConfig) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	for i := range rr.endpoints {
		if rr.endpoints[i].URL == endpoint.URL {
			rr.endpoints[i].Healthy = true
			break
		}
	}
}

// MarkUnhealthy marks an endpoint as unhealthy
func (rr *RoundRobin) MarkUnhealthy(endpoint *models.EndpointConfig) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	for i := range rr.endpoints {
		if rr.endpoints[i].URL == endpoint.URL {
			rr.endpoints[i].Healthy = false
			break
		}
	}
}

// Weighted implements weighted round-robin load balancing
type Weighted struct {
	endpoints      []models.EndpointConfig
	weightedList   []int
	current        uint32
	mutex          sync.RWMutex
}

// NewWeighted creates a new weighted load balancer
func NewWeighted(endpoints []models.EndpointConfig) *Weighted {
	w := &Weighted{
		endpoints: endpoints,
	}
	w.buildWeightedList()
	return w
}

// buildWeightedList builds the weighted list of endpoint indices
func (w *Weighted) buildWeightedList() {
	w.weightedList = make([]int, 0)
	for i, ep := range w.endpoints {
		if ep.Healthy {
			for j := 0; j < ep.Weight; j++ {
				w.weightedList = append(w.weightedList, i)
			}
		}
	}
}

// Next returns the next endpoint based on weights
func (w *Weighted) Next() *models.EndpointConfig {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if len(w.weightedList) == 0 {
		return nil
	}

	index := atomic.AddUint32(&w.current, 1) % uint32(len(w.weightedList))
	endpointIndex := w.weightedList[index]
	return &w.endpoints[endpointIndex]
}

// MarkHealthy marks an endpoint as healthy
func (w *Weighted) MarkHealthy(endpoint *models.EndpointConfig) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for i := range w.endpoints {
		if w.endpoints[i].URL == endpoint.URL {
			w.endpoints[i].Healthy = true
			w.buildWeightedList()
			break
		}
	}
}

// MarkUnhealthy marks an endpoint as unhealthy
func (w *Weighted) MarkUnhealthy(endpoint *models.EndpointConfig) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for i := range w.endpoints {
		if w.endpoints[i].URL == endpoint.URL {
			w.endpoints[i].Healthy = false
			w.buildWeightedList()
			break
		}
	}
}

// LeastConnections implements least connections load balancing
type LeastConnections struct {
	endpoints   []models.EndpointConfig
	connections map[string]int32
	mutex       sync.RWMutex
}

// NewLeastConnections creates a new least connections load balancer
func NewLeastConnections(endpoints []models.EndpointConfig) *LeastConnections {
	lc := &LeastConnections{
		endpoints:   endpoints,
		connections: make(map[string]int32),
	}

	for _, ep := range endpoints {
		lc.connections[ep.URL] = 0
	}

	return lc
}

// Next returns the endpoint with least connections
func (lc *LeastConnections) Next() *models.EndpointConfig {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	var selected *models.EndpointConfig
	minConnections := int32(^uint32(0) >> 1) // Max int32

	for i := range lc.endpoints {
		ep := &lc.endpoints[i]
		if !ep.Healthy {
			continue
		}

		conn := lc.connections[ep.URL]
		if conn < minConnections {
			minConnections = conn
			selected = ep
		}
	}

	if selected != nil {
		lc.connections[selected.URL]++
	}

	return selected
}

// MarkHealthy marks an endpoint as healthy
func (lc *LeastConnections) MarkHealthy(endpoint *models.EndpointConfig) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	for i := range lc.endpoints {
		if lc.endpoints[i].URL == endpoint.URL {
			lc.endpoints[i].Healthy = true
			break
		}
	}
}

// MarkUnhealthy marks an endpoint as unhealthy
func (lc *LeastConnections) MarkUnhealthy(endpoint *models.EndpointConfig) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	for i := range lc.endpoints {
		if lc.endpoints[i].URL == endpoint.URL {
			lc.endpoints[i].Healthy = false
			break
		}
	}
}

// Random implements random load balancing
type Random struct {
	endpoints []models.EndpointConfig
	mutex     sync.RWMutex
}

// NewRandom creates a new random load balancer
func NewRandom(endpoints []models.EndpointConfig) *Random {
	return &Random{
		endpoints: endpoints,
	}
}

// Next returns a random healthy endpoint
func (r *Random) Next() *models.EndpointConfig {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Find healthy endpoints
	healthyEndpoints := make([]models.EndpointConfig, 0)
	for _, ep := range r.endpoints {
		if ep.Healthy {
			healthyEndpoints = append(healthyEndpoints, ep)
		}
	}

	if len(healthyEndpoints) == 0 {
		return nil
	}

	// Random selection (simplified - should use proper random)
	index := int(atomic.AddUint32(&randomSeed, 1)) % len(healthyEndpoints)
	return &healthyEndpoints[index]
}

// MarkHealthy marks an endpoint as healthy
func (r *Random) MarkHealthy(endpoint *models.EndpointConfig) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i := range r.endpoints {
		if r.endpoints[i].URL == endpoint.URL {
			r.endpoints[i].Healthy = true
			break
		}
	}
}

// MarkUnhealthy marks an endpoint as unhealthy
func (r *Random) MarkUnhealthy(endpoint *models.EndpointConfig) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i := range r.endpoints {
		if r.endpoints[i].URL == endpoint.URL {
			r.endpoints[i].Healthy = false
			break
		}
	}
}

var randomSeed uint32