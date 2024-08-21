package main

import (
	"bufio"
	"fmt"
	"os"

	. "bvh-anim-parser/types"
	"bvh-anim-parser/utils"

	// import vec3, mat4 and quatetnion from go3d
	"github.com/ungerik/go3d/float64/quaternion"
	"github.com/ungerik/go3d/float64/vec3"
)

// Opens and parses a BVH file.
func ParseBVH(filepath string) (*BVH, error) {

	//// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//// Create a scanner to read the file. By default it reads line by line.
	scanner := bufio.NewScanner(file)

	//// Partially initialize BVH struct. It will be fully initialized later.
	var bvh *BVH = &BVH{
		Joints:   []*Joint{},
		EndSites: []*Joint{},
	}
	
	//// Parse the hierarchy
	err = utils.ParseHierarchy(scanner, bvh)
	if err != nil {
		return nil, err
	}

	//// Initialize Pose
	for _, joint := range bvh.Joints {
		joint.Pose.GlobalPos = make([]vec3.T, bvh.NumFrames)
		joint.Pose.GlobalRot = make([]quaternion.T, bvh.NumFrames)
		joint.Pose.PosOffsetFromRest = make([]vec3.T, bvh.NumFrames)
		joint.Pose.RotOffsetFromRest = make([]quaternion.T, bvh.NumFrames)
	}

	//// check all rotations are of the same order
	var root = bvh.Joints[0]
	bvh.RotationOrder = string(root.Channels[3][0]) + string(root.Channels[4][0]) + string(root.Channels[5][0])
	for _, joint := range bvh.Joints[1:] {
		currentRotationOrder := string(joint.Channels[0][0]) + string(joint.Channels[1][0]) + string(joint.Channels[2][0])
		if currentRotationOrder != bvh.RotationOrder {
			return nil, fmt.Errorf("invalid rotation order for joint %s: found %v, should be %v", joint.Name, currentRotationOrder, bvh.RotationOrder)
		}
	}

	//// Parse the motion
	err = utils.ParseMotion(scanner, bvh)
	if err != nil {
		return nil, err
	}
	

	//// Calculate RestPose (GlobalPosition, GlobalRotation, FromParentRotation) for all joints
	for _, joint := range bvh.Joints {
		utils.CalcRestPose(joint)
	}

	//// Calculate Pose (GlobalPosition, GlobalRotation, FromRestPosition) for all joints for each frame
	for _, joint := range bvh.Joints {
		for frame := 0; frame < bvh.NumFrames; frame++ {
			utils.CalcPose(joint, frame)
		}
	}

	return bvh, nil
}

func main() {
	bvhPath := "./test_anim.bvh"

	bvh, err := ParseBVH(bvhPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Parsed BVH File: %+v\n", bvh)

	bvhJSON, err := utils.ToJSON(bvh)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	// save the JSON to a file
	jsonFile, err := os.Create("bvh.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer jsonFile.Close()
	
	jsonFile.WriteString(bvhJSON)
	fmt.Println("BVH JSON saved to bvh.json")
}
