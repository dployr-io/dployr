#!/usr/bin/env bats

source "$BATS_TEST_DIRNAME/../../install.sh"

@test "render_dployrd_config: contains base_url" {
  run render_dployrd_config "https://base.dployr.io" "inst-abc" "instance" "" "" "0" "0" "0"
  [[ "$output" == *'base_url = "https://base.dployr.io"'* ]]
}

@test "render_dployrd_config: contains instance_id" {
  run render_dployrd_config "https://base.dployr.io" "inst-abc" "instance" "" "" "0" "0" "0"
  [[ "$output" == *'instance_id = "inst-abc"'* ]]
}

@test "render_dployrd_config: contains node_role" {
  run render_dployrd_config "https://base.dployr.io" "inst-abc" "build" "" "" "0" "0" "0"
  [[ "$output" == *'node_role = "build"'* ]]
}

@test "render_dployrd_config: contains registry_url and registry_auth" {
  run render_dployrd_config "https://base.dployr.io" "inst-abc" "build" \
    "registry.digitalocean.com/myapp" "do-token-xyz" "0" "0" "0"
  [[ "$output" == *'registry_url = "registry.digitalocean.com/myapp"'* ]]
  [[ "$output" == *'registry_auth = "do-token-xyz"'* ]]
}

@test "render_dployrd_config: contains resource limits" {
  run render_dployrd_config "https://base.dployr.io" "inst-abc" "instance" "" "" "512" "2" "10"
  [[ "$output" == *'container_memory = 512'* ]]
  [[ "$output" == *'container_cpu = 2'* ]]
  [[ "$output" == *'container_storage = 10'* ]]
}

@test "render_dployrd_config: binds to localhost port 7879" {
  run render_dployrd_config "https://base.dployr.io" "inst-abc" "instance" "" "" "0" "0" "0"
  [[ "$output" == *'address = "localhost"'* ]]
  [[ "$output" == *'port = 7879'* ]]
}

@test "render_sudoers: contains systemctl rule" {
  run render_sudoers "/bin/systemctl" "/sbin/reboot" "/bin/mkdir" "/bin/rm" \
    "/bin/cp" "/bin/chmod" "/usr/bin/tee" "/usr/bin/docker"
  [[ "$output" == *"dployrd ALL=(ALL) NOPASSWD: /bin/systemctl *"* ]]
}

@test "render_sudoers: contains reboot rule" {
  run render_sudoers "/bin/systemctl" "/sbin/reboot" "/bin/mkdir" "/bin/rm" \
    "/bin/cp" "/bin/chmod" "/usr/bin/tee" "/usr/bin/docker"
  [[ "$output" == *"dployrd ALL=(ALL) NOPASSWD: /sbin/reboot"* ]]
}

@test "render_sudoers: contains docker rule" {
  run render_sudoers "/bin/systemctl" "/sbin/reboot" "/bin/mkdir" "/bin/rm" \
    "/bin/cp" "/bin/chmod" "/usr/bin/tee" "/usr/bin/docker"
  [[ "$output" == *"dployrd ALL=(ALL) NOPASSWD: /usr/bin/docker *"* ]]
}

@test "render_sudoers: contains all eight binary rules" {
  run render_sudoers "/bin/systemctl" "/sbin/reboot" "/bin/mkdir" "/bin/rm" \
    "/bin/cp" "/bin/chmod" "/usr/bin/tee" "/usr/bin/docker"
  local count
  count=$(echo "$output" | grep -c "NOPASSWD")
  [ "$count" -eq 8 ]
}

@test "parse_json: extracts a top-level string field" {
  command -v jq >/dev/null 2>&1 || skip "jq not available"
  result=$(echo '{"tag_name":"v1.2.3"}' | parse_json '.tag_name')
  [ "$result" = "v1.2.3" ]
}

@test "parse_json: returns empty string for missing field" {
  command -v jq >/dev/null 2>&1 || skip "jq not available"
  result=$(echo '{}' | parse_json '.tag_name')
  [ "$result" = "" ]
}

@test "parse_json: returns empty string for null value" {
  command -v jq >/dev/null 2>&1 || skip "jq not available"
  result=$(echo '{"tag_name":null}' | parse_json '.tag_name')
  [ "$result" = "" ]
}
