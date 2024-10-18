## No longer being worked on
When I set out to build this, it was out of a mix of need and frustration. I was spending a lot of time rendering audio down in Ableton and then uploading it to Soundcloud to reference on speakers outside of my studio space. This app effectively eliminated the need to upload the audio to Soundcloud. Soon after reaching this point in development, I discovered (Sonobus)[https://sonobus.net/], a free, open-source, fully featured network audio streaming application that then eliminated my need to render audio down at all. I promptly abandoned this project. 

I've added a very liberal license. If you feel you can benefit from what is here in any way, please be my guest.

### Overview
I’m working on a Golang application that serves music files over HTTP, allowing users to access these files from outside the network using ngrok. Users can add multiple directory paths that the app will serve, and multiple ngrok domains that the app will use to tunnel the local server.

### **Usefule Links and Resouces**
https://ngrok.com/docs/api/#quickstart
https://pkg.go.dev/github.com/ngrok/ngrok-api-go/v4
https://api.beatport.com/v4/docs/
https://developers.soundcloud.com/docs/api/guide
https://bandcamp.com/developer



### **Technologies In Use:**

- **Golang**: Core application logic and HTTP file serving.
- **ngrok**: Used to expose the local server to the internet.
- **JSON Configuration**: Storing user preferences for directories and ngrok domains persistently in `config.json`.
- **HTML/CSS**: Simple web interface for displaying music files and allowing users to add new directories and ngrok domains.

### **Accomplished So Far:**

- The app serves music files from multiple directories over HTTP.
- Users can configure multiple ngrok domains for external access, and the app automatically runs ngrok on those domains at startup.
- A simple web interface allows users to manage directory paths and ngrok domains.
- All configurations (directories and domains) persist across sessions via a local configuration file (`config.json`).
- Error handling is in place to check for missing directories or ngrok domains and prompt the user accordingly.

### **Potential Future Features:**

- **User Authentication**: Implement basic username/password authentication to protect access to the music files.
- **HTTPS Support**: Secure the connection by enabling HTTPS, especially when exposing the app via ngrok.
- **File Uploads**: Add support for users to upload files directly through the web interface.
- **Advanced File Browsing**: Provide a more robust UI for navigating directories, such as displaying folder trees or adding file filters.
- **Dynamic Port Selection**: Allow users to select a different port for serving the HTTP server, especially if `:80` is unavailable.
- **Playlist Support**: Implement a feature for users to create and manage playlists from the served music files.
- **File Streaming**: Improve the file streaming experience by adding features like progress bars or an embedded media player.
  
  ### **Current Bugs**
  Site does not include an actual media player, so songs don't play on my mobile device
