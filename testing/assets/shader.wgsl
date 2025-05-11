struct VertexOutput {
//    @location(0) tex_coord: vec2<f32>,
    @builtin(position) position: vec4<f32>,
};

@group(0)
@binding(0)
var<uniform> transform: mat4x4<f32>;

@vertex
fn vs_main(
    @location(0) position: vec4<f32>,
//    @location(1) tex_coord: vec2<f32>,
) -> VertexOutput {
    var myMatrix: mat4x4<f32> = mat4x4<f32>(
    1.7342978, -0.34566143, -0.27681828, -0.24913645, // column 0
    0.5202893, 1.1522048, 0.92272764, 0.8304548, // column 1
    0.0, 2.0931718, -0.55363655, -0.4982729, // column 2
    0.0, 0.0, 5.5786643, 6.0207977  // column 3
    );

var myMatrix2: mat4x4<f32> = mat4x4<f32>(
    vec4<f32>(1.0, 0.0, 0.0, 0.0),
    vec4<f32>(0.0, 1.0, 0.0, 0.0),
    vec4<f32>(0.0, 0.0, 1.0, 0.0),
    vec4<f32>(0.0, 0.0, 0.0, 1.0)
);
myMatrix2 = myMatrix2 * transform;

    var result: VertexOutput;
//    result.tex_coord = tex_coord;
    result.position = myMatrix * position;
    return result;
}

//@group(0)
//@binding(1)
//var r_color: texture_2d<u32>;

@fragment
fn fs_main(vertex: VertexOutput) -> @location(0) vec4<f32> {
//    let tex = textureLoad(r_color, vec2<i32>(vertex.tex_coord * 256.0), 0);
//    let v = f32(tex.x) / 255.0;
    let v = 1.0 / 255.0;
//    return vec4<f32>(1.0 - (v * 5.0), 1.0 - (v * 15.0), 1.0 - (v * 50.0), 1.0);
    return vec4<f32>(255.0, 255.0, 255.0, 255.0);
}

@fragment
fn fs_wire(vertex: VertexOutput) -> @location(0) vec4<f32> {
    return vec4<f32>(0.0, 0.5, 0.0, 0.5);
}
