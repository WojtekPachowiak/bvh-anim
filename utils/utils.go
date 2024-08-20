package utils

import (
	"bufio"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/ungerik/go3d/float64/quaternion"
	"github.com/ungerik/go3d/float64/vec3"

	. "bvh-anim-parser/types"
)

// create method for quaternion that takes x, y, and z angles in degrees AND rotation order and then returns the quaternion
func QuaternionFromEulerAngles(x, y, z float64, order string) quaternion.T {
	// convert to radians
	x = x * math.Pi / 180
	y = y * math.Pi / 180
	z = z * math.Pi / 180

	qx := quaternion.FromXAxisAngle(x)
	qy := quaternion.FromYAxisAngle(y)
	qz := quaternion.FromZAxisAngle(z)

	switch order {
	case "XYZ":
		return quaternion.Mul3(&qz, &qy, &qx)
	case "XZY":
		return quaternion.Mul3(&qy, &qz, &qx)
	case "YXZ":
		return quaternion.Mul3(&qz, &qx, &qy)
	case "YZX":
		return quaternion.Mul3(&qx, &qz, &qy)
	case "ZXY":
		return quaternion.Mul3(&qy, &qx, &qz)
	case "ZYX":
		return quaternion.Mul3(&qx, &qy, &qz)
	default:
		panic("[QuaternionFromEulerAngles] BUG! Invalid rotation order.")
	}
}

// Each joint can be interpreted as a bone with a head and a tail. This function calculates the offset from head to tail.
func GetTailOffset(joint *Joint) vec3.T {
	if joint.IsEndSite {
		return joint.RestPose.PosOffsetFromParent
	}

	// calculate average of child positions
	var sum vec3.T
	for _, child := range joint.Children {
		sum = vec3.Add(&sum, &child.RestPose.PosOffsetFromParent)
	}
	return *sum.Scale(1.0 / float64(len(joint.Children)))
}

func ParseHierarchy(scanner *bufio.Scanner, bvh *BVH) error {
	var depth int = 0
	var currentJoint *Joint

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)

		// JOINT RightUpLeg
		if fields[0] == "ROOT" || fields[0] == "JOINT" {
			isRoot := fields[0] == "ROOT"
			currentJoint = &Joint{
				Name:      fields[1],
				Depth:     depth,
				IsEndSite: false,
				IsRoot:    isRoot,
				IsLeaf:    false,
				Children:  []*Joint{},
				Parent:    currentJoint,
				RestPose:  &RestPose{},
				Pose:      &Pose{},
			}
			if !isRoot {
				currentJoint.Parent.Children = append(currentJoint.Parent.Children, currentJoint)
			}
			bvh.Joints = append(bvh.Joints, currentJoint)

			// End Site
		} else if strings.HasPrefix(strings.ToLower(line), "end") {
			currentJoint.IsLeaf = true
			var endSite *Joint = &Joint{
				Name:      "ENDSITE",
				Depth:     depth,
				IsEndSite: true,
				IsLeaf:    false,
				Children:  nil,
				RestPose: &RestPose{
					PosOffsetFromParent: vec3.T{0, 0, 0},
				},
				Pose:     nil,
				Channels: nil,
				IsRoot:   false,
				Parent:   currentJoint,
			}
			endSite.Parent.Children = append(endSite.Parent.Children, endSite)
			bvh.EndSites = append(bvh.EndSites, endSite)

		} else if fields[0] == "{" {
			depth++

		} else if fields[0] == "}" {
			depth--

			// OFFSET 45.6091 0 0
		} else if fields[0] == "OFFSET" {
			offset := strings.Fields(line)
			if len(offset) < 4 {
				return fmt.Errorf("invalid OFFSET line: %s", line)
			}
			x, _ := strconv.ParseFloat(offset[1], 64)
			y, _ := strconv.ParseFloat(offset[2], 64)
			z, _ := strconv.ParseFloat(offset[3], 64)
			currentJoint.RestPose.PosOffsetFromParent = vec3.T{x, y, z}

			// Channels 3 Zrotation Xrotation Yrotation
		} else if fields[0] == "CHANNELS" {
			numChannels, _ := strconv.Atoi(fields[1])
			if currentJoint.IsRoot && numChannels != 6 {
				return fmt.Errorf("invalid number of channels for root joint: %s", line)
			} else if !currentJoint.IsRoot && numChannels != 3 {
				return fmt.Errorf("invalid number of channels for non-root joint: %s", line)
			}

			currentJoint.Channels = fields[2:]
			bvh.NumAllChannels += numChannels

			// Frames: 129
		} else if fields[0] == "Frames:" {
			if len(fields) != 2 {
				return fmt.Errorf("invalid number of frames: %s", line)
			}
			bvh.NumFrames, _ = strconv.Atoi(fields[1])

			// Frame time: 0.033333
		} else if fields[0] == "Frame" {
			bvh.FrameTime, _ = strconv.ParseFloat(strings.Fields(line)[2], 64)
			bvh.Fps = 1 / bvh.FrameTime
			break // we are done parsing the hierarchy, now we can parse the motion data
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func ParseMotion(scanner *bufio.Scanner, bvh *BVH) error {
	var frame = 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		fieldsf := make([]float64, len(fields))

		if len(fields) != bvh.NumAllChannels {
			return fmt.Errorf("invalid number of channels in frame %d: found %d, should be %d", frame, len(fields), bvh.NumAllChannels)
		}

		// convert all fields to float64
		for i, field := range fields {
			fieldsf[i], _ = strconv.ParseFloat(field, 64)
		}

		for i, joint := range bvh.Joints {
			if joint.IsRoot {
				joint.Pose.GlobalPos = append(joint.Pose.GlobalPos, vec3.T{fieldsf[0], fieldsf[1], fieldsf[2]})
			}
			quat := QuaternionFromEulerAngles(fieldsf[3+3*i], fieldsf[4+3*i], fieldsf[5+3*i], bvh.RotationOrder)
			joint.Pose.RotOffsetFromRest = append(joint.Pose.RotOffsetFromRest, quat)
		}

		frame++

	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if frame != bvh.NumFrames {
		return fmt.Errorf("invalid number of frames: found %d, should be %d", frame, bvh.NumFrames)
	}

	return nil
}

// TODO: verify correctness
func CalcPose(joint *Joint, frame int) {
	if joint.IsRoot {
		joint.Pose.GlobalRot[frame] = joint.Pose.RotOffsetFromRest[frame]

		joint.Pose.PosOffsetFromRest[frame] = vec3.Sub(
			&joint.Pose.GlobalPos[frame],
			&joint.RestPose.GlobalPos)

		// joint.Pose.GlobalPosition[frame] has already been calculated when parsing MOTION section
	} else {
		var rotatedOffset vec3.T = joint.Pose.RotOffsetFromRest[frame].RotatedVec3(&joint.RestPose.PosOffsetFromParent)

		joint.Pose.GlobalPos[frame] = vec3.Add(
			&joint.Parent.Pose.GlobalPos[frame],
			&rotatedOffset)

		joint.Pose.GlobalRot[frame] = quaternion.Mul(
			&joint.Parent.Pose.GlobalRot[frame],
			&joint.Pose.RotOffsetFromRest[frame])

		joint.Pose.PosOffsetFromRest[frame] = vec3.Sub(
			&rotatedOffset,
			&joint.RestPose.PosOffsetFromParent)
	}
}

func CalcRestPose(joint *Joint) {
	//// RestPose.GlobalPosition
	if joint.IsRoot {
		joint.RestPose.GlobalPos = joint.RestPose.PosOffsetFromParent
	} else {
		joint.RestPose.GlobalPos = vec3.Add(&joint.Parent.RestPose.GlobalPos, &joint.RestPose.PosOffsetFromParent)
	}

	//// RestPose.GlobalRotation
	var tailOffset vec3.T = GetTailOffset(joint)
	var dir vec3.T
	if !tailOffset.IsZero() {
		dir = tailOffset
		dir.Normalize()
	} else {
		dir = vec3.T{0, 1, 0}
	}
	var axs vec3.T = vec3.T{0, 1, 0}
	var dot float64 = vec3.Dot(&axs, &dir)
	var restGlobalRotation quaternion.T
	if dot < -0.9999 {
		restGlobalRotation = quaternion.T{0, 0, 0, 1}
	} else if dot > 0.9999 {
		restGlobalRotation = quaternion.T{1, 0, 0, 0}
	} else {
		var angle float64 = math.Acos(dot)
		var axis vec3.T = vec3.Cross(&axs, &dir)
		axis.Normalize()
		restGlobalRotation = quaternion.FromAxisAngle(&axis, angle)
	}
	joint.RestPose.GlobalRot = restGlobalRotation

	//// RestPose.FromParentRotation (TODO: verify that the order of multiplication is correct)
	if joint.IsRoot {
		joint.RestPose.RotOffsetFromParent = joint.RestPose.GlobalRot
	} else {
		parentQuatInv := joint.Parent.RestPose.GlobalRot.Inverted()
		joint.RestPose.RotOffsetFromParent = quaternion.Mul(&parentQuatInv, &joint.RestPose.GlobalRot)
	}
}
