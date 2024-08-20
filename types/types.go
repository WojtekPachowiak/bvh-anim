package types

import (
	"github.com/ungerik/go3d/float64/quaternion"
	"github.com/ungerik/go3d/float64/vec3"
)

// BVH represents the entire BVH file.
type BVH struct {
	Fps           float64 // frames per second
	Joints        []*Joint
	EndSites      []*Joint
	NumFrames     int     // number of frames in the animation
	FrameTime     float64 // the same as 1/Fps
	RotationOrder string  // the order of rotation for all joints (e.g. "XYZ", "YXZ", etc.)
	NumAllChannels int    // the total number of channels in the BVH file
}

// Joint represents a joint in the BVH hierarchy.
type Joint struct {
	Name      string
	Parent    *Joint
	Children  []*Joint
	Depth     int
	IsEndSite bool
	IsLeaf    bool
	IsRoot    bool
	Channels  []string
	RestPose  *RestPose
	Pose      *Pose
}

// RestPose represents the initial position and orientation of a joint (frame independent).
type RestPose struct {
	GlobalPos           vec3.T
	GlobalRot           quaternion.T
	RotOffsetFromParent quaternion.T // The rotation that takes the parent orientation to the current orientation.
	PosOffsetFromParent vec3.T       // This is the same thing as OFFSET in HIERARCHY.
}

// Pose represents the current position and orientation of a joint at a given frame.
type Pose struct {
	GlobalPos         []vec3.T       // Position with respect to the world space (origin).
	GlobalRot         []quaternion.T // Rotation with respect to the world space (basis).
	PosOffsetFromRest []vec3.T       // Also known as local position.
	RotOffsetFromRest []quaternion.T // Also known as local rotation.
}
