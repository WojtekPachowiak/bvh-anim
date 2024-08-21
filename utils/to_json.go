package utils

import (
	. "bvh-anim-parser/types"
	"encoding/json"
	"fmt"
)

// BVH represents the entire BVH file.
type BVHSerializable struct {
	Fps            float64 // frames per second
	Joints         []*JointSerializable
	EndSites       []*JointSerializable
	NumFrames      int     // number of frames in the animation
	FrameTime      float64 // the same as 1/Fps
	RotationOrder  string  // the order of rotation for all joints (e.g. "XYZ", "YXZ", etc.)
	NumAllChannels int     // the total number of channels in the BVH file
}

// Joint represents a joint in the BVH hierarchy.
type JointSerializable struct {
	Name      string
	Parent    string
	Children  []string
	Depth     int
	IsEndSite bool
	IsLeaf    bool
	IsRoot    bool
	Channels  []string
	RestPose  *RestPose
	Pose      *Pose
}

func ToJSON(bvh *BVH) (string, error) {

	// convert BVH to Dict
	bvhS := &BVHSerializable{
		Fps:            bvh.Fps,
		Joints:         []*JointSerializable{},
		EndSites:       []*JointSerializable{},
		NumFrames:      bvh.NumFrames,
		FrameTime:      bvh.FrameTime,
		RotationOrder:  bvh.RotationOrder,
		NumAllChannels: bvh.NumAllChannels,
	}

	for _, joint := range bvh.Joints {
		jointS := &JointSerializable{
			Name:      joint.Name,
			Parent:     "",
			Children:  []string{},
			Depth:     joint.Depth,
			IsEndSite: joint.IsEndSite,
			IsLeaf:    joint.IsLeaf,
			IsRoot:    joint.IsRoot,
			Channels:  joint.Channels,
			RestPose:  joint.RestPose,
			Pose:      joint.Pose,
		}
		if joint.Parent != nil {
			jointS.Parent = joint.Parent.Name
		}
		for _, child := range joint.Children {
			jointS.Children = append(jointS.Children, child.Name)
		}

		bvhS.Joints = append(bvhS.Joints, jointS)
	}

	for _, endSite := range bvh.EndSites {
		endSiteDict := &JointSerializable{
			Name:      endSite.Name,
			Parent:    endSite.Parent.Name,
			Children:  []string{},
			Depth:     endSite.Depth,
			IsEndSite: endSite.IsEndSite,
			IsLeaf:    endSite.IsLeaf,
			IsRoot:    endSite.IsRoot,
			Channels:  endSite.Channels,
			RestPose:  endSite.RestPose,
			Pose:      endSite.Pose,
		}

		bvhS.EndSites = append(bvhS.EndSites, endSiteDict)
	}

	bvhJSON, err := json.MarshalIndent(bvhS, "", "    ")
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(bvhJSON), nil
}
