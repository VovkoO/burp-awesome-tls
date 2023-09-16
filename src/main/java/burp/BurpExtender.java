package burp;

import com.google.gson.Gson;
import com.sun.jna.Native;

import java.io.PrintWriter;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;

public class BurpExtender implements IBurpExtender, IHttpListener, IExtensionStateListener, IProxyListener {
    private PrintWriter stdout;
    private PrintWriter stderr;
    private Gson gson;
    private Settings settings;

    private IExtensionHelpers helpers;
    private IBurpExtenderCallbacks callbacks;

    private static final String HEADER_KEY = "Awesometlsconfig";

    @Override
    public void registerExtenderCallbacks(IBurpExtenderCallbacks callbacks)
    {
        this.stdout = new PrintWriter(callbacks.getStdout(), true);
        this.stderr = new PrintWriter(callbacks.getStderr(), true);
        this.callbacks = callbacks;
        this.helpers = callbacks.getHelpers();
        this.gson = new Gson();
        this.settings = new Settings(callbacks);

        callbacks.setExtensionName("Awesome TLS");
        callbacks.registerHttpListener(this);
        callbacks.registerProxyListener(this);
        callbacks.registerExtensionStateListener(this);
        callbacks.addSuiteTab(new SettingsTab(this.settings));

        new Thread(() -> {
<<<<<<< Updated upstream
//             this.stdout.println(Platform.RESOURCE_PREFIX);

            this.stdout.println("load");
            var err = ServerLibrary.INSTANCE.StartServer(
                this.settings.getInterceptProxyAddress(),
                this.settings.getBurpProxyAddress(),
                this.settings.getSpoofProxyAddress());

            this.stdout.println("loaded");

=======
            System.setProperty("jna.library.path", "/Users/vladimir.razdrogin/GolandProjects/burp-awesome-tls/build/resources/main/darwin-aarch64/server.so");
            this.stdout.println("LOAD");

            try {
                ServerLibrary INSTANCE = Native.load("server", ServerLibrary.class);
            } catch (UnsatisfiedLinkError e) {
                this.stdout.println("Не удалось найти или загрузить библиотеку: " + e);
            }

            var err = ServerLibrary.INSTANCE.StartServer(this.settings.getAddress());
            this.stdout.println("LOADED");
>>>>>>> Stashed changes
            if (!err.equals("")) {
                var isGraceful = err.contains("Server stopped"); // server was stopped gracefully by calling StopServer()
                var out = isGraceful ? this.stdout : this.stderr;
                out.println(err);
                if (!isGraceful) callbacks.unloadExtension(); // fatal error; disable the extension
            }
        }).start();
    }

    @Override
    public void processHttpMessage(int toolFlag, boolean messageIsRequest, IHttpRequestResponse messageInfo) {
        if (!messageIsRequest) return;

        var httpService = messageInfo.getHttpService();
        var req = this.helpers.analyzeRequest(messageInfo.getRequest());

        var transportConfig = new TransportConfig();
        transportConfig.Host = httpService.getHost();
        transportConfig.Scheme = httpService.getProtocol();
        transportConfig.Fingerprint = this.settings.getFingerprint();
        transportConfig.HexClientHello = this.settings.getHexClientHello();
        transportConfig.HttpTimeout = this.settings.getHttpTimeout();
        transportConfig.HttpKeepAliveInterval = this.settings.getHttpKeepAliveInterval();
        transportConfig.IdleConnTimeout = this.settings.getIdleConnTimeout();
        transportConfig.TlsHandshakeTimeout = this.settings.getTlsHandshakeTimeout();
        transportConfig.UseInterceptedFingerprint = this.settings.getUseInterceptedFingerprint();
        var goConfigJSON = this.gson.toJson(transportConfig);
        this.stdout.println("Using config: " + goConfigJSON);

        var headers = req.getHeaders();
        headers.add(HEADER_KEY + ": " + goConfigJSON);

        try {
            var url = new URL("https://" + this.settings.getSpoofProxyAddress());
            messageInfo.setHttpService(helpers.buildHttpService(url.getHost(), url.getPort(), url.getProtocol()));
            messageInfo.setRequest(helpers.buildHttpMessage(headers, Arrays.copyOfRange(messageInfo.getRequest(), req.getBodyOffset(), messageInfo.getRequest().length)));
        } catch (Exception e) {
            this.stderr.println("Failed to intercept http service: " + e);
            this.callbacks.unloadExtension();
        }
    }

    @Override
    public void extensionUnloaded() {
        var err = ServerLibrary.INSTANCE.StopServer();
        if (!err.equals("")) {
            this.stderr.println(err);
        }
    }

    @Override
    public void processProxyMessage(boolean messageIsRequest, IInterceptedProxyMessage message) {
        if (message.getMessageInfo().getHttpService().getHost().equals("awesome-tls-error")) {
            var bodyOffset = this.helpers.analyzeRequest(message.getMessageInfo().getRequest()).getBodyOffset();
            var req = message.getMessageInfo().getRequest();
            var body = Arrays.copyOfRange(req, bodyOffset, req.length);
            this.stderr.println(new String(body, StandardCharsets.UTF_8));
        }
    }
}
