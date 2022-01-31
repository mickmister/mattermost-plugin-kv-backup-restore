// This file is automatically generated. Do not modify it manually.

const manifest = JSON.parse(`
{
    "id": "kv-backup-restore",
    "name": "KV Store Backup/Restore",
    "description": "This plugin helps save the state of a given plugin's kvstore values.",
    "version": "0.2.0",
    "min_server_version": "5.24.0",
    "server": {
        "executables": {
            "linux-amd64": "server/dist/plugin-linux-amd64",
            "darwin-amd64": "server/dist/plugin-darwin-amd64",
            "windows-amd64": "server/dist/plugin-windows-amd64.exe"
        },
        "executable": ""
    },
    "webapp": {
        "bundle_path": "webapp/dist/main.js"
    },
    "settings_schema": {
        "header": "",
        "footer": "",
        "settings": []
    }
}
`);

export default manifest;
export const id = manifest.id;
export const version = manifest.version;
