import asyncore
import socket


        

class HTTPClient(asyncore.dispatcher):

    def __init__(self, parent,data):
        asyncore.dispatcher.__init__(self)
        self.create_socket(socket.AF_INET, socket.SOCK_STREAM)
        self.connect( ("127.0.0.1", 4007) )
        self.buffer = data
        self.parent=parent
        print 'connet;'

    def handle_connect(self):
        pass

    def handle_close(self):
        self.parent.close()
        try:
            self.close()
        except:
            pass

    def handle_read(self):
        print 'data'
        self.parent.sendall(self.recv(8192))

    def writable(self):
        return (len(self.buffer) > 0)

    def handle_write(self):
        sent = self.send(self.buffer)
        self.buffer = self.buffer[sent:]
globvar = {}
class EchoHandler(asyncore.dispatcher_with_send):
    #self.data =None
    #self.cli = None

    def handle_close(self):
        global globvar
        print 'close'
        try:
            globvar[str(self.getpeername())].close()
        except:
            pass 
        try:
            del globvar[str(self.getpeername())]
        except:
            pass
        
        self.close()
    

    def handle_read(self):
        global globvar
        data = self.recv(8192)
        if data .lower().find("policy-file-request") != -1:
            s1= '<cross-domain-policy><allow-access-from domain="*" to-ports="*" /></cross-domain-policy>\0' #.encode('utf-8')+0
            #s2="""HTTP/1.1 200 OK\r\nContent-Length: {len}\r\nContent-Type: text/xml\r\n\r\n""".replace("{len}",str(len(s1)))+s1

            self.sendall(s1)#'<?xml version="1.0"?><cross-domain-policy><allow-access-from domain="*" to-ports="*" /></cross-domain-policy>')
            #self.data=""
            return
        #if data.lower().find("crossdomain.xml")>=0 and self.data.find("\r\n\r\n")<0:           return
        elif data.lower().find("crossdomain.xml") != -1:
            #print data
            s1= '<?xml version="1.0"?><cross-domain-policy><allow-access-from domain="*" to-ports="*" /></cross-domain-policy>'
            s2="""HTTP/1.1 200 OK\r\nContent-Length: {len}\r\nContent-Type: text/xml\r\n\r\n""".replace("{len}",str(len(s1)))+s1
            self.sendall(s2)#         self.data = ""
            return

        if  data:
            if globvar.has_key(str(self.getpeername()))== False:
                globvar[str(self.getpeername())] = HTTPClient(self,data)
            else:
                globvar[str(self.getpeername())].send(data) #lf.data=""
        
class EchoServer(asyncore.dispatcher):

    def __init__(self, host, port):
        asyncore.dispatcher.__init__(self)
        self.create_socket(socket.AF_INET, socket.SOCK_STREAM)
        self.set_reuse_addr()
        self.bind((host, port))
        self.listen(1024)

    def handle_accept(self):
        pair = self.accept()
        if pair is not None:
            sock, addr = pair
            print 'Incoming connection from %s' % repr(addr)
            handler = EchoHandler(sock)

server = EchoServer('0.0.0.0', 60007)
asyncore.loop()
