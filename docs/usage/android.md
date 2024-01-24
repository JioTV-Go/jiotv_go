# Guide to Run JioTV Go on Android

This guide will walk you through the process of downloading the Go binary from GitHub Releases and running it on your Android device using two different methods: Termux and UserLand.

## Method 1: Using Termux

### Step 1: Install Termux

Download latest version of Termux installed from [F-Droid](https://f-droid.org/en/packages/com.termux/) or [GitHub Release](https://github.com/termux/termux-app/releases/latest).

Play store version of Termux is not recommended as it is outdated and can cause issues.

### Step 2: Install Dependencies

1. Open Termux.
2. Run the following command to update pkg index:

```bash
pkg update
```

This will take some time.

3. Next, Install `wget` and `openssl` by running the following commands:

```bash
pkg install curl openssl
```

<div class="warning">

> After this, we recommend following the easy method, use [automatic install script](../get_started.md#automatic-installation-recommended) to install JioTV Go on Android.

</div>

If you want to install manually, follow the steps below:

### Step 3: Download JioTV Go Binary

4. Identify your architecture by running the following command:

```bash
uname -m
```

Now look for the architecture in the [table](./get_started.md#the-following-table-lists-the-binaries-available-for-download). Long press on binary name and click on `Copy Link Address`.

4. Paste the link in the following command and run it:

```bash
curl -LO jiotv_go --progress-bar "https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-android-{arch}"
```

If you are using above URL, replace `{arch}` with your architecture.

### Step 4: Make the binary executable and run it

5. Make the binary executable by running the following command:

```bash
chmod +x jiotv_go
```

6. Execute the binary by running the following command:

```bash
./jiotv_go [commands]
```

See the [Usage](../usage/usage.md) page for more information about commands.

If you want to run from any directory, without `./` prefix, then you can move the binary to `$PREFIX/bin` directory by running the following command:

```bash
mv jiotv_go $PREFIX/bin
```

Now you can run the binary from any directory without `./` prefix.

```bash
jiotv_go [commands]
```


## Method 2: Using UserLand (Linux Environment)

### Step 1: Install UserLand

1. Download UserLand from [Google Play Store](https://play.google.com/store/apps/details?id=tech.ula) or [F-Droid](https://f-droid.org/en/packages/tech.ula/) and install it on your Android device.

2. Open UserLand and choose Alpine with Terminal.

3. Wait for the setup to complete.

### Step 2: Install Dependencies

1. Install `wget` and `openssl` by running the following commands:

```bash
apk update && apk add curl
```

<div class="warning">

> Remember UserLand has a linux environment, hence you need to use linux binaries.
> 
> After this, we recommend following the easy method, use [automatic install script](../get_started.md#automatic-installation-recommended) to install JioTV Go on Linux.
>
> If you want to install manually, follow the steps below:

</div>

### Step 3: Download JioTV Go Binary

2. Identify your architecture by running the following command:

```bash
uname -m
```

Now look for the linux binaries with architecture in the [table](../get_started.md#the-following-table-lists-the-binaries-available-for-download). Long press on binary name and click on `Copy Link Address`. Do not use Android binaries, as they will not work in UserLand.

3. Paste the link in the following command and run it:

```bash
curl -LO jiotv_go --progress-bar "https://github.com/rabilrbl/jiotv_go/releases/latest/download/jiotv_go-linux-{arch}"
```

If you are using above URL, replace `{arch}` with your architecture.

### Step 4: Make the binary executable and run it

4. Make the binary executable by running the following command:

```bash
chmod +x jiotv_go
```

5. Execute the binary by running the following command:

```bash
./jiotv_go [commands]
```

See the [Usage](../usage/usage.md) page for more information about commands.

If you want to run from any directory, without `./` prefix, then you can move the binary to `/usr/bin` directory by running the following command:

```bash
mv jiotv_go /usr/bin
```

Now you can run the binary from any directory without `./` prefix.

```bash
jiotv_go [commands]
```

## Conclusion:

You've successfully installed and run JioTV Go on your Android device using either Termux or UserLand. Enjoy your favorite TV shows and channels!
