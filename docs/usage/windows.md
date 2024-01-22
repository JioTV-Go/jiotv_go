# Using JioTV Go on Windows

After you have downloaded the latest release of JioTV Go, you can run it on Windows by following these steps:

1. Open terminal in the directory where you have downloaded JioTV Go.
   
   This can be done by opening the folder where you have downloaded JioTV Go, and then typing `cmd` in the address bar.

   Or you can open the folder where you have downloaded JioTV Go, and then press `Shift + Right Click` and then click on `Open Terminal Here`.

   Or you can also open Windows Terminal and then type `cd` followed by a space, and then drag and drop the folder where you have downloaded JioTV Go.

2. Run the following command:

   ```sh
    jiotv_go-windows-{arch}.exe serve
    ```

    Replace `{arch}` with your architecture. For example, if your architecture is `amd64`, then you will run the following command:

    ```sh
    jiotv_go-windows-amd64.exe serve
    ```

    If you are unsure about your architecture, read the [Identify your architecture](../get_started.md#identifying-your-os-and-architecture) section in the [Get Started](../get_started.md) page.