# New DRM Component

We have implemented a new DRM component for the web which enables new channels that JioTV added recently.

**Important Considerations:**

*   **IPTV Client Compatibility:** This new DRM component hasn't been implemented to work with an IPTV client at the moment. This is the primary reason it's not ready for a full release.
*   **Pre-release Access:** If you are impatient to try it, you can check our recent pre-release on GitHub. Docker users can use the `develop` tag in their image URL.
*   **HTTPS Requirement:** DRM will only work on `http://localhost` or `https`. This means if you're not using `localhost` or `127.0.0.1` as your host, you **must** enable https (TLS 1.3) using self-signed certificates. We have enabled this by default in JioTV Go. Users who have a reverse proxy with https won't be affected and can use the application as before with DRM enabled.
*   **Non-Technical Users:** For non-technical users, please be aware that DRM will only work if you are running JioTV Go on the same device you are viewing from. Cross-device access (e.g., running on a server and viewing on a different computer/phone) won't be possible when DRM is enabled.
*   **Supported Browsers:** Only official browsers such as Firefox and Chrome are supported from our end due to DRM (Widevine L3). It might work on other browsers, but we won't provide support for them.
*   **Community Support:** You are free to discuss this on the community support group, but please **do not** bother admins and moderators with how-to questions regarding this experimental feature.
*   **Stable Release:** The stable release will be delayed until we decide on the support for IPTV clients with the new DRM component.

See [config.md](./config.md#drm-digital-rights-management) for how to enable DRM.
