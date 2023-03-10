import Layout from "./components/layout/Layout";
import { AuthProvider } from "./context/AuthContext";
import { AttachmentProvider } from "./context/AttachmentContext";
import { InterfaceProvider } from "./context/InterfaceContext";
import { ModalProvider } from "./context/ModalContext";
import { MouseProvider } from "./context/MouseContext";
import { SocketProvider } from "./context/SocketContext";
import { UserdropdownProvider } from "./context/UserdropdownContext";
import { UsersProvider } from "./context/UsersContext";

function App() {
  return (
    <InterfaceProvider>
      <MouseProvider>
        <SocketProvider>
          <ModalProvider>
            <UserdropdownProvider>
              <AuthProvider>
                <AttachmentProvider>
                  <UsersProvider>
                    <Layout />
                  </UsersProvider>
                </AttachmentProvider>
              </AuthProvider>
            </UserdropdownProvider>
          </ModalProvider>
        </SocketProvider>
      </MouseProvider>
    </InterfaceProvider>
  );
}

export default App;
