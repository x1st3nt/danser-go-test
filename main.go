package main

/*
#ifdef _WIN32
#include <windows.h>
// force switch to the high performance gpu in multi-gpu systems (mostly laptops)
__declspec(dllexport) DWORD NvOptimusEnablement = 0x00000001; // http://developer.download.nvidia.com/devzone/devcenter/gamegraphics/files/OptimusRenderingPolicies.pdf
__declspec(dllexport) DWORD AmdPowerXpressRequestHighPerformance = 0x00000001; // https://community.amd.com/thread/169965
#endif
*/
import "C"

import (
	"fmt"
	"log"
	"os"
	
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/wieku/danser-go/app"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/launcher"
)

var leftClickCounter, rightClickCounter int

func main() {
	// Initialize environment
	env.Init("danser")

	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatalln("Failed to initialize GLFW:", err)
	}
	defer glfw.Terminate()

	// Create a windowed mode window and its OpenGL context
	window, err := glfw.CreateWindow(800, 600, "Key Counter Overlay", nil, nil)
	if err != nil {
		log.Fatalln("Failed to create GLFW window:", err)
	}
	window.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		log.Fatalln("Failed to initialize OpenGL:", err)
	}

	// Set callbacks for key and mouse input
	window.SetKeyCallback(keyCallback)
	window.SetMouseButtonCallback(mouseButtonCallback)

	if len(os.Args) == 1 {
		launcher.StartLauncher()
	} else {
		go app.Run() // Start the app in a goroutine
	}

	// Main loop for the overlay
	for !window.ShouldClose() {
		// Render overlay here
		render()

		// Swap front and back buffers
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		// Increment key counter (example for space key)
		if key == glfw.KeySpace {
			leftClickCounter++
		}
	}
}

func mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		switch button {
		case glfw.MouseButtonLeft:
			leftClickCounter++
		case glfw.MouseButtonRight:
			rightClickCounter++
		}
	}
}

func render() {
	// Clear the screen
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// Render text showing counts (implement text rendering here)
	fmt.Printf("Left Clicks: %d, Right Clicks: %d\n", leftClickCounter, rightClickCounter)

	// Add text rendering logic here to display on the overlay
}
