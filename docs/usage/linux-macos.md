# Using JioTV Go on Linux and macOS

After you have downloaded the latest release of JioTV Go, you can run it on Linux and macOS by following these steps:

1. Open terminal.
2. Navigate to the directory where you have downloaded JioTV Go.
3. Make the file executable by running the following command:

   ```sh
   chmod +x jiotv_go-linux-{arch}
   ```

   Replace `{arch}` with your architecture. For example, if your architecture is `amd64`, then you will run the following command:

   ```sh
   chmod +x jiotv_go-linux-amd64
   ```

   If you are unsure about your architecture, read the [Identify your architecture](../get_started.md#identifying-your-os-and-architecture) section in the [Get Started](../get_started.md) page.

4. Run the following command:

   ```sh
    ./jiotv_go-linux-{arch} serve
    ```

5. Open your web browser and visit [http://localhost:5001/](http://localhost:5001/).
