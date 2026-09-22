// name_conversion_test.cpp — regression test for the carried upstream patches in
// patches/ that touch src/name_conversion.cpp. Compiled against a patched
// upstream tree by scripts/test-upstream-patches.sh (no ggml backend, no model
// files needed).
#include <cstdio>
#include <string>
#include <vector>

#include "name_conversion.h"

// The only core/util.cpp helpers name_conversion.cpp links against. Defined here
// (same semantics as upstream) so the test doesn't have to link util.cpp and,
// through it, ggml.
bool ends_with(const std::string& str, const std::string& ending) {
    return str.size() >= ending.size() && str.compare(str.size() - ending.size(), ending.size(), ending) == 0;
}

bool starts_with(const std::string& str, const std::string& start) {
    return str.compare(0, start.size(), start) == 0;
}

bool contains(const std::string& str, const std::string& substr) {
    return str.find(substr) != std::string::npos;
}

std::vector<std::string> split_string(const std::string& str, char delimiter) {
    std::vector<std::string> result;
    size_t start = 0;
    size_t end   = str.find(delimiter);
    while (end != std::string::npos) {
        result.push_back(str.substr(start, end - start));
        start = end + 1;
        end   = str.find(delimiter, start);
    }
    result.push_back(str.substr(start));
    return result;
}

static int failures = 0;

static void expect(const char* in, SDVersion version, const char* want) {
    std::string got = convert_tensor_name(in, version);
    if (got != want) {
        std::fprintf(stderr, "FAIL convert_tensor_name(\"%s\", %d)\n  got:  %s\n  want: %s\n", in, (int)version, got.c_str(), want);
        failures++;
    }
}

int main() {
    // SD3.x all-in-one files with legacy cond_stage_model.* text encoder prefixes
    // (e.g. gpustack/stable-diffusion-v3-5-large-turbo-GGUF) map to the
    // text_encoders.* slots SD3CLIPEmbedder loads.
    expect("cond_stage_model.transformer.text_model.encoder.layers.0.mlp.fc1.weight", VERSION_SD3,
           "text_encoders.clip_l.transformer.text_model.encoder.layers.0.mlp.fc1.weight");
    expect("cond_stage_model.transformer.text_model.embeddings.token_embedding.weight", VERSION_SD3,
           "text_encoders.clip_l.transformer.text_model.embeddings.token_embedding.weight");
    expect("cond_stage_model.1.transformer.text_model.encoder.layers.31.self_attn.q_proj.bias", VERSION_SD3,
           "text_encoders.clip_g.transformer.text_model.encoder.layers.31.self_attn.q_proj.bias");
    expect("cond_stage_model.1.transformer.text_model.text_projection", VERSION_SD3,
           "text_encoders.clip_g.transformer.text_model.text_projection");
    expect("cond_stage_model.2.transformer.encoder.block.0.layer.0.SelfAttention.relative_attention_bias.weight", VERSION_SD3,
           "text_encoders.t5xxl.transformer.encoder.block.0.layer.0.SelfAttention.relative_attention_bias.weight");
    expect("cond_stage_model.2.transformer.encoder.block.23.layer.1.DenseReluDense.wi_0.weight", VERSION_SD3,
           "text_encoders.t5xxl.transformer.encoder.block.23.layer.1.DenseReluDense.wi_0.weight");
    expect("cond_stage_model.2.transformer.encoder.embed_tokens.weight", VERSION_SD3,
           "text_encoders.t5xxl.transformer.encoder.embed_tokens.weight");
    expect("cond_stage_model.2.transformer.shared.weight", VERSION_SD3,
           "text_encoders.t5xxl.transformer.shared.weight");

    // Already-canonical SD3 names are untouched.
    expect("text_encoders.clip_g.transformer.text_model.final_layer_norm.weight", VERSION_SD3,
           "text_encoders.clip_g.transformer.text_model.final_layer_norm.weight");
    expect("text_encoders.t5xxl.transformer.encoder.final_layer_norm.weight", VERSION_SD3,
           "text_encoders.t5xxl.transformer.encoder.final_layer_norm.weight");

    // SD1.x / SDXL legitimately use cond_stage_model.*: unchanged.
    expect("cond_stage_model.transformer.text_model.encoder.layers.0.mlp.fc1.weight", VERSION_SD1,
           "cond_stage_model.transformer.text_model.encoder.layers.0.mlp.fc1.weight");
    expect("cond_stage_model.transformer.text_model.encoder.layers.0.mlp.fc1.weight", VERSION_SDXL,
           "cond_stage_model.transformer.text_model.encoder.layers.0.mlp.fc1.weight");
    expect("cond_stage_model.1.transformer.text_model.text_projection", VERSION_SDXL,
           "cond_stage_model.1.transformer.text_model.text_projection");

    // FLUX: unchanged.
    expect("cond_stage_model.transformer.text_model.encoder.layers.0.mlp.fc1.weight", VERSION_FLUX,
           "cond_stage_model.transformer.text_model.encoder.layers.0.mlp.fc1.weight");
    expect("text_encoders.clip_l.transformer.text_model.encoder.layers.0.mlp.fc1.weight", VERSION_FLUX,
           "text_encoders.clip_l.transformer.text_model.encoder.layers.0.mlp.fc1.weight");

    if (failures) {
        std::fprintf(stderr, "%d failure(s)\n", failures);
        return 1;
    }
    std::printf("name_conversion_test: ok\n");
    return 0;
}
