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
/*
var pcConfig = {
    'iceServers': [{
        'urls': 'stun:stun.l.google.com:19302'
    }],
};
*/

var pcConfig = null;

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

    socket = new WebSocket("wss://" + document.location.host + "/r/" + room);
    socket.onmessage = function (event) {
        var data = JSON.parse(event.data);

        var clientId = null;
        if (data.sub.con.clientId){
            clientId = data.sub.con.clientId;
        }
        console.log("clientId: " + clientId);
        var msg = JSON.parse(data.data);
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
                    console.log("New Peer: Create Offer");
                    pc = new RTCPeerConnection(pcConfig);
                    pc.onremovestream = handleRemoteStreamRemoved;
                    pc.onnegotiationneeded = handleOnNegotiationNeeded;
                    pc.onicegatheringstatechange = handleIceGatheringStateChange;
                    pc.oniceconnectionstatechange = handleIceConnectionStateChange;
                    pc.ontrack = handleRemoteTrack;
                    localStream.getTracks().forEach(track => {
                        console.log("New Track");
                        pc.addTrack(track, localStream);
                    });
                    pc.onicecandidate = function(event) {
                        if (event.candidate) {
                            console.log("Local Candidate: ", event.candidate);                            
                            /*
                            socket.send(JSON.stringify({
                                type: 'candidate',
                                label: event.candidate.sdpMLineIndex,
                                id: event.candidate.sdpMid,
                                candidate: event.candidate.candidate,
                                clientId: clientId
                            }));
                            */
                        } else {
                            socket.send(JSON.stringify({type: "offer", sdp: pc.localDescription.sdp, clientId: clientId}));
                        }
                    };
                    pc.createOffer().then(
                        function(desc){
                            console.log("Offer: set local Description");
                            if (connections.has(clientId)){
                                connection = connections.get(clientId);
                                connection.setLocalDescription(desc);
                                //socket.send(JSON.stringify({type: desc.type, sdp: desc.sdp, clientId: clientId}));
                            } else {
                                console.log("Client not in Map");
                            }
                        }, onCreateSessionDescriptionError
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
                //if(clientId !== null && msg.clientId == localClientId){    
                    
                    if(msg.clientId && msg.clientId !== null && msg.clientId != "") {
                        clientId = msg.clientId;
                    }
                    
                    pc = null;
                    if (connections.has(clientId)){
                        console.log("Existing Peer");
                        pc = connections.get(clientId);
                    }
                    else if (connections.size < remoteVideo.length) {
                        console.log("New Peer");
                        pc = new RTCPeerConnection(pcConfig);
                        if(clientId == localClientId){
                            localStream.getTracks().forEach(track => {
                                console.log("New Track");
                                pc.addTrack(track, localStream);
                            });
                        }
                        connections.set(clientId, pc);
                    }
                    if(pc !== null){
                        console.log("Peer Connection is not null");
                        pc.onremovestream = handleRemoteStreamRemoved;
                        pc.onnegotiationneeded = handleOnNegotiationNeeded;
                        pc.onicegatheringstatechange = handleIceGatheringStateChange;
                        pc.oniceconnectionstatechange = handleIceConnectionStateChange;
                        pc.ontrack = handleRemoteTrack;
                        pc.onicecandidate = function(event) {
                            if (event.candidate) {
                                console.log("Local Candidate: ", event.candidate);             
                                /*
                                socket.send(JSON.stringify({
                                    type: 'candidate',
                                    label: event.candidate.sdpMLineIndex,
                                    id: event.candidate.sdpMid,
                                    candidate: event.candidate.candidate,
                                    clientId: clientId
                                }));
                                */
                            } else {
                                socket.send(JSON.stringify({type: "answer", sdp: pc.localDescription.sdp, clientId: clientId}));
                            }
                        };
                        pc.setRemoteDescription(new RTCSessionDescription(msg));
                        pc.createAnswer().then(
                            function(desc){
                                console.log("Answer: set local Description");
                                if (connections.has(clientId)){
                                    connection = connections.get(clientId);
                                    connection.setLocalDescription(desc);
                                    //socket.send(JSON.stringify({type: desc.type, sdp: desc.sdp, clientId: clientId}));
                                } else {
                                    console.log("Client not in Map");
                                }
                            }, onCreateSessionDescriptionError
                        );
                    }
                //}
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
                console.log("Remote Candidate: ", msg);
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
            candidate: event.candidate.candidate,
            clientId: clientId
        }));
    } else {
        console.log('End of candidates.');
    }
}

function handleRemoteTrack(event) {
    console.log("Got Remote Track");
    console.log(event);
    if(event.track.kind == "video"){
        var index = remoteStream.length;
        remoteStream.push(event.streams[0]);
        remoteVideo[index].srcObject = event.streams[0];
        remoteVideo[index].autoplay = true;
    }
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

function hangup() {
    console.log('Hanging up.');
    socket.send(JSON.stringify({ type: 'message', message: 'bye' }));
}

function handleIceConnectionStateChange(event) {
    console.log("iceconnectionstatechanged event: ", event);
}

function handleIceGatheringStateChange(event) {
    console.log("icegatheringstatechanged event: ", event);
}

function handleCreateOfferError(event) {
    console.log('createOffer() error: ', event);
}

function onCreateSessionDescriptionError(error) {
    console.log('Failed to create session description: ' + error.toString());
}

function handleOnNegotiationNeeded(event){
    console.log("Negotiation is needed. Event: ", event)
}

function handleRemoteStreamRemoved(event) {
    console.log('Remote stream removed. Event: ', event);
}



