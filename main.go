package main

import (
	"context"
	"log"
	"net"
	"os"
	"path"
	"time"

	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	resourceName = "github.com/mosvector"
	devicePath   = "/dev/mosvector"
	socketPath   = pluginapi.DevicePluginPath + "mosvector.sock"
)

type MosvectorDevicePlugin struct {
	server *grpc.Server
}

func NewMosvectorDevicePlugin() *MosvectorDevicePlugin {
	log.Println("Creating new MosvectorDevicePlugin instance.")
	return &MosvectorDevicePlugin{}
}

// Start the device plugin
func (plugin *MosvectorDevicePlugin) Start() error {
	log.Println("Starting MosvectorDevicePlugin...")
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Error removing socket file: %v", err)
		return err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Printf("Error creating listener on socket: %v", err)
		return err
	}

	plugin.server = grpc.NewServer()
	pluginapi.RegisterDevicePluginServer(plugin.server, plugin)

	go func() {
		log.Println("Serving gRPC server for MosvectorDevicePlugin...")
		if err := plugin.server.Serve(listener); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for the server to start by checking the socket
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(socketPath); err == nil {
			log.Println("Socket found; server is up and running.")
			break
		}
		log.Printf("Waiting for socket to be created (attempt %d)...", i+1)
		time.Sleep(time.Second)
	}
	return plugin.registerWithKubelet()
}

// Stop the device plugin
func (plugin *MosvectorDevicePlugin) Stop() {
	log.Println("Stopping MosvectorDevicePlugin...")
	if plugin.server != nil {
		plugin.server.Stop()
		plugin.server = nil
		log.Println("gRPC server stopped.")
	}
	if err := os.Remove(socketPath); err != nil {
		log.Printf("Error removing socket file during stop: %v", err)
	} else {
		log.Println("Socket file removed.")
	}
}

// Register with kubelet
func (plugin *MosvectorDevicePlugin) registerWithKubelet() error {
	log.Println("Registering MosvectorDevicePlugin with Kubelet...")
	timeout := 5 * time.Second
	conn, err := grpc.Dial(pluginapi.KubeletSocket, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)
	if err != nil {
		log.Printf("Error connecting to Kubelet: %v", err)
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(socketPath),
		ResourceName: resourceName,
	}

	if _, err := client.Register(context.Background(), req); err != nil {
		log.Printf("Error registering device plugin: %v", err)
		return err
	}

	log.Println("Successfully registered MosvectorDevicePlugin with Kubelet.")
	return nil
}

// ListAndWatch lists devices and updates the list on changes.
func (plugin *MosvectorDevicePlugin) ListAndWatch(req *pluginapi.Empty, srv pluginapi.DevicePlugin_ListAndWatchServer) error {
	log.Println("ListAndWatch called. Sending initial device list...")
	devices := []*pluginapi.Device{
		{ID: "mosvector-0", Health: pluginapi.Healthy},
	}

	if err := srv.Send(&pluginapi.ListAndWatchResponse{Devices: devices}); err != nil {
		log.Printf("Failed to send initial device list: %v", err)
	} else {
		log.Println("Initial device list sent successfully.")
	}

	for {
		log.Println("ListAndWatch: Sending updated device health status...")
		if err := srv.Send(&pluginapi.ListAndWatchResponse{Devices: devices}); err != nil {
			log.Printf("Error sending updated device list: %v", err)
		}
		time.Sleep(5 * time.Second)
	}
}

// Allocate assigns the device to the requesting pod.
func (plugin *MosvectorDevicePlugin) Allocate(ctx context.Context, req *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	log.Printf("Allocate called with %d container requests.", len(req.ContainerRequests))
	responses := &pluginapi.AllocateResponse{}

	for _, containerReq := range req.ContainerRequests {
		log.Printf("Allocating device for container request: %v", containerReq.DevicesIDs)
		response := pluginapi.ContainerAllocateResponse{
			Devices: []*pluginapi.DeviceSpec{
				{HostPath: devicePath, ContainerPath: devicePath, Permissions: "rw"},
			},
		}
		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}

	log.Println("Devices allocated successfully.")
	return responses, nil
}

// GetDevicePluginOptions provides options for the device plugin.
func (plugin *MosvectorDevicePlugin) GetDevicePluginOptions(ctx context.Context, empty *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	log.Println("GetDevicePluginOptions called.")
	return &pluginapi.DevicePluginOptions{}, nil
}

// GetPreferredAllocation provides a preferred device allocation, used when multiple devices are requested.
func (plugin *MosvectorDevicePlugin) GetPreferredAllocation(ctx context.Context, req *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	log.Println("GetPreferredAllocation called.")
	return &pluginapi.PreferredAllocationResponse{}, nil
}

// PreStartContainer is called before a container starts.
func (plugin *MosvectorDevicePlugin) PreStartContainer(ctx context.Context, req *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	log.Println("PreStartContainer called.")
	return &pluginapi.PreStartContainerResponse{}, nil
}

func main() {
	log.Println("Starting MosvectorDevicePlugin main process...")
	plugin := NewMosvectorDevicePlugin()

	if err := plugin.Start(); err != nil {
		log.Fatalf("Error starting device plugin: %v", err)
	}
	defer plugin.Stop()

	log.Println("MosvectorDevicePlugin is running. Waiting indefinitely...")
	// Run forever
	select {}
}
