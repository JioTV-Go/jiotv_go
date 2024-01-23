# Running JioTV Go on Linux and macOS

Once you have downloaded the [latest release of JioTV Go](../get_started.md#pre-built-binaries), follow these steps to run it seamlessly on Linux and macOS:

1. **Open Terminal:**
   - Launch the terminal on your Linux or macOS system.

2. **Navigate to Downloaded Directory:**
   - Move to the directory where you have downloaded JioTV Go.

3. **Make the File Executable:**
   - Execute the following command to make the file executable:
   
     ```sh
     chmod +x jiotv_go-linux-{arch}
     ```

     Replace `{arch}` with your architecture. For example, if your architecture is `amd64`, use the command:

     ```sh
     chmod +x jiotv_go-linux-amd64
     ```

     If you are unsure about your architecture, check the [Identify your architecture](../get_started.md#identifying-your-os-and-architecture) section in the [Get Started](../get_started.md) page.

4. **Run JioTV Go:**
   - Start JioTV Go by running the following command:

     ```sh
     ./jiotv_go-linux-{arch} serve
     ```

5. **Access the Server:**
   - Open your web browser and go to [http://localhost:5001/](http://localhost:5001/) to access JioTV Go.

Enjoy your JioTV Go streaming experience on Linux and macOS! If you encounter any issues or have questions, refer to the [Support and Issues](#support-and-issues) section in the user guide or visit the [GitHub repository](https://github.com/rabilrbl/jiotv_go).