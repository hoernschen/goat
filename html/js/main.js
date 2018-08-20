'use strict';
var room = 'foo';
var localClientId;
var isChannelReady = false;
var isInitiator = false;
var socket;
var localStream;
var pc;
var connection;
var connections = new Map();
var remoteStream = [];
var turnReady;

var pcConfig = {
    'iceServers': [{
        'urls': 'stun:stun.l.google.com:19302'
    }]
};

var constraints = {
    video: true,
    audio: false
};

/////////////////////////////////////////////

var localVideo = document.querySelector('#localVideo');
var remoteVideo = document.querySelectorAll('.remoteVideo');
navigator.mediaDevices.getUserMedia(constraints)
.then(function(stream) {
    localStream = stream;
    localVideo.srcObject = stream;

    socket = new WebSocket("wss://" + document.location.host + "/ws/" + room);
    socket.onmessage = function (event) {
        console.log(event);
        var data = JSON.parse(event.data);
        console.log(data);
        var clientId = null;
        if (data.sub.con.clientId){
            clientId = data.sub.con.clientId;
        }
        console.log(clientId);
        var msg = JSON.parse(data.data);
        console.log(msg);
        console.log(msg.type);
        if(msg.Type == 1){
            msg.type = "offer";
        }
        switch (msg.type) {
            case "created":
                console.log('Created room ' + msg.text);
                isInitiator = true;
                if (clientId !== null) {
                    localClientId = clientId;
                    console.log("local Client Id: " + localClientId);
                }
                break;
            case "join":
                isChannelReady = true;
                if (connections.size < remoteVideo.length && clientId !== null) {
                    console.log("New Peer: Send Offer");
                    pc = new RTCPeerConnection(null);
                    pc.onicecandidate = handleIceCandidate;
                    pc.onremovestream = handleRemoteStreamRemoved;
                    pc.ontrack = handleRemoteTrack;
                    localStream.getTracks().forEach(track => {
                        pc.addTrack(track, localStream);
                    });
                    pc.createOffer(
                        function (desc) {
                            console.log("Offer: set local Description");
                            pc.setLocalDescription(desc);
                            socket.send(JSON.stringify(desc));
                        }, handleCreateOfferError
                    );
                    connections.set(clientId,pc);
                }
                break;
            case "joined":
                isChannelReady = true;
                if(clientId !== null){
                    localClientId = clientId;
                    console.log("local Client Id: " + localClientId);
                }
                break;
            case "offer":
                console.log("Got Offer");
                if(clientId !== null){    
                    
                    pc = null;
                    if (connections.has(clientId)){
                        pc = connections.get(clientId);
                    }
                    else if (connections.size < remoteVideo.length) {
                        pc = new RTCPeerConnection(null);
                        pc.onicecandidate = handleIceCandidate;
                        pc.onremovestream = handleRemoteStreamRemoved;
                        pc.ontrack = handleRemoteTrack;
                        localStream.getTracks().forEach(track => {
                            pc.addTrack(track, localStream);
                        });
                        connections.set(clientId, pc);
                    }
                    if(pc !== null){
                        pc.setRemoteDescription(new RTCSessionDescription(msg));
                        pc.createAnswer().then(
                            function(desc){
                                console.log("Answer: set local Description");
                                if (connections.has(clientId)){
                                    connection = connections.get(clientId);
                                    connection.setLocalDescription(desc);
                                    socket.send(JSON.stringify({type: desc.type, sdp: desc.sdp, clientId: clientId}));
                                } else {
                                    console.log("Client not in Map");
                                }
                            }, onCreateSessionDescriptionError
                        );
                    }
                }
                break;
            case "answer":
                console.log("Got Answer");
                if (connections.has(clientId) && msg.clientId == localClientId) {
                    delete msg.clientId;
                    connection = connections.get(clientId);
                    connection.setRemoteDescription(new RTCSessionDescription(msg));
                }
                break;
            case "candidate":
                console.log("candidate");
                connection = connections.get(clientId);
                if (connection && connection !== null){
                    var candidate = new RTCIceCandidate({
                        sdpMLineIndex: msg.label,
                        candidate: msg.candidate
                    });
                    connection.addIceCandidate(candidate);
                }
                break;
            default:
                if (msg.text === "bye" && clientId !== null) {
                    handleRemoteHangup(clientId);
                }
        }
    }
})
.catch(function (e) {
    console.log('getUserMedia() error: ' + e.name);
});

window.onbeforeunload = function () {
    socket.send(JSON.stringify({ type: 'message', message: 'bye' }));
};

/////////////////////////////////////////////////////////

function handleIceCandidate(event) {
    console.log('icecandidate event: ', event);
    if (event.candidate) {
        socket.send(JSON.stringify({
            type: 'candidate',
            label: event.candidate.sdpMLineIndex,
            id: event.candidate.sdpMid,
            candidate: event.candidate.candidate
        }));
    } else {
        console.log('End of candidates.');
    }
}

function handleCreateOfferError(event) {
    console.log('createOffer() error: ', event);
}

function onCreateSessionDescriptionError(error) {
    console.log('Failed to create session description: ' + error.toString());
}

function handleRemoteTrack(event) {
    console.log("Got Remote Track");
    if(event.track.kind == "video"){
        var index = remoteStream.length;
        remoteStream.push(event.streams[0]);
        remoteVideo[index].srcObject = event.streams[0];
    }
}

function handleRemoteStreamRemoved(event) {
    console.log('Remote stream removed. Event: ', event);
}

function hangup() {
    console.log('Hanging up.');
    socket.send(JSON.stringify({ type: 'message', message: 'bye' }));
}

function handleRemoteHangup(clientId) {
    console.log('Session terminated.');
    connection = connections[clientId];
    if(connection && connection !== null){
        connection.close();
        connections.delete(clientId);
    }
    isInitiator = false;
}