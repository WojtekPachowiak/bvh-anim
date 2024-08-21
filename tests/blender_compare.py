import bpy

# load bvh file
bpy.ops.import_anim.bvh(
    filepath="C:/Users/.../bvh_files/01_01.bvh"
    )
bvh = bpy.context.object


# enter edit mode
bpy.ops.object.mode_set(mode='EDIT')


s,e = bvh.animation_data.action.frame_range
info = {
    'num_frames': round(e-s),
    'joints': {bone.name: {} for bone in bvh.data.edit_bones},
}


# collect rest pose data
for bone in bvh.data.edit_bones:
    rest_pose = {
        'global_pos': tuple(bone.matrix.to_translation()),
        'global_rot': tuple(bone.matrix.to_quaternion()),
        'rot_offset_from_parent': tuple(bone.matrix.to_quaternion() * bone.parent.matrix.to_quaternion().inverted()),
        'pos_offset_from_parent': tuple(bone.head - bone.parent.head) if bone.parent else tuple(bone.head),
    }
    info['joints'][bone.name]['rest_pose'] = rest_pose
    

# enter pose mode
bpy.ops.object.mode_set(mode='POSE')


# collect pose data
for bone in bvh.pose.bones:
    pose = {
        'global_pos': [],
        'global_rot': [],
        'pos_offset_from_rest': [],
        'rot_offset_from_rest': [],
    }
    for f in range(round(s), round(e) +1):
        bpy.context.scene.frame_set(f)
        pose['global_pos'].append(tuple(bone.matrix.to_translation()))
        pose['global_rot'].append(tuple(bone.matrix.to_quaternion()))
        pose['pos_offset_from_rest'].append(tuple(bone.location))
        pose['rot_offset_from_rest'].append(tuple(bone.rotation_quaternion))
        
    info['joints'][bone.name]['pose'] = pose
    
    
# save info to json file
import json
with open("tests/blender_compare.json", 'w') as f:
    json.dump(info, f, indent=4)