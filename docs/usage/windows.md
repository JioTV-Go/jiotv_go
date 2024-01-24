# Using JioTV Go on Windows

Welcome to JioTV Go, your gateway to a seamless TV streaming experience on Windows! Follow these straightforward steps to get started without delving into technical complexities:

<div class="warning">

> Use [automatic install script](../get_started.md#windows) to install JioTV Go on Windows.
>
> If you want to install manually, follow the steps below:

</div>

## Manual Installation

We assume that you have already downloaded the [latest release of JioTV Go](../get_started.md#pre-built-binaries) for Windows. If not, please download it before proceeding.

### Step 1: Open JioTV Go Folder

Locate the folder where you downloaded JioTV Go. Here are three easy ways to open the terminal in that directory:

- **Option 1:** Open the folder and type `cmd` in the address bar.
- **Option 2:** Right-click the folder while holding `Shift` and select "Open Terminal Here."
- **Option 3:** Open Windows Terminal, type `cd`, add a space, and then drag and drop the JioTV Go folder into the terminal.

### Step 2: Launch JioTV Go

Run the following command in the terminal:

```powershell
.\jiotv_go-windows-{arch}.exe serve
```

Remember to replace `{arch}` with your architecture. For example, if your architecture is `x86_64`, use the following command:

```powershell
.\jiotv_go-windows-amd64.exe serve
```

Unsure about your architecture? Check out the [Identify your architecture](../get_started.md#identifying-your-os-and-architecture) section on the [Get Started](../get_started.md) page for guidance.

That's it! You're now ready to enjoy JioTV Go on your Windows system hassle-free. If you encounter any issues or have questions, refer to our user-friendly documentation or reach out to our support team for assistance. Happy streaming!