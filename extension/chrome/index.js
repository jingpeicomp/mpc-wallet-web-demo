const wasmUrl = chrome.runtime.getURL('main.wasm');

const go = new Go();
WebAssembly.instantiateStreaming(fetch(wasmUrl), go.importObject).then((result) => {
    go.run(result.instance);
    console.log('Key value is ', keyNum)
});


let partyNo1;
let fileKey1Bytes;
let partyNo2;
let fileKey2Bytes;

function signAction() {
    let inputMsg = $("#inputMsg").val();
    if (!inputMsg) {
        $("#alertModalBody").text("请输入待签名的信息");
        $('#alertModal').modal('show')
        return;
    }

    $("#signBtnStatus").show();
    let signResult = sign(partyNo1, fileKey1Bytes, partyNo2, fileKey2Bytes, inputMsg);
    console.log("-----> Sign finish, result : ", signResult)
    $("#signBtnStatus").hide();
    if (signResult) {
        $('#signResult').val(signResult)
    }
}

function parseKeyAction() {
    let addrAndPriKey = parseKey();
    if (addrAndPriKey) {
        $('#walletAddress').val(addrAndPriKey[0])
        $('#walletPriKey').val(addrAndPriKey[1])
    }
}

function setKeyAction() {
    if (!fileKey1Bytes || !fileKey2Bytes) {
        $("#alertModalBody").text("请上传密钥分片");
        $('#alertModal').modal('show')
        return;
    }

    setKey(partyNo1, fileKey1Bytes, partyNo2, fileKey2Bytes);
    fileKey1Bytes = null;
    fileKey2Bytes = null;
    $("#fileInput1Label").text("密钥分片1已保存");
    $("#fileInput2Label").text("密钥分片2已保存");
}

function clearKeyAction() {
    clearKey()
    $("#fileInput1Label").text("请上传密钥分片1");
    $("#fileInput2Label").text("请上传密钥分片2");
}

$(document).ready(function () {
    $("#signBtn").click(function () {
        console.log('Sign button is clicked')
        signAction();
    });

    $("#viewBtn").click(function () {
        console.log('Parse key button is clicked')
        parseKeyAction();
    });

    $("#setKeyBtn").click(function () {
        console.log('Set key button is clicked')
        setKeyAction();
    });

    $("#clearKeyBtn").click(function () {
        console.log('Clear key button is clicked')
        clearKeyAction();
    });
});

window.onload = function (e) {
    document.getElementById('fileInput1')
        .addEventListener('change', function selectedFileChanged() {
            if (this.files.length === 0) {
                console.log('请选择文件！');
                return;
            }

            if (this.files[0].name) {
                $("#fileInput1Label").text(this.files[0].name);
                partyNo1 = parseInt(this.files[0].name.substring(3))
            }
            const reader = new FileReader();
            reader.onload = function fileReadCompleted() {
                fileKey1Bytes = new Uint8Array(reader.result);
            };
            reader.readAsArrayBuffer(this.files[0]);
        });

    document.getElementById('fileInput2')
        .addEventListener('change', function selectedFileChanged() {
            if (this.files.length === 0) {
                console.log('请选择文件！');
                return;
            }

            if (this.files[0].name) {
                $("#fileInput2Label").text(this.files[0].name);
                partyNo2 = parseInt(this.files[0].name.substring(3))
            }
            const reader = new FileReader();
            reader.onload = function fileReadCompleted() {
                fileKey2Bytes = new Uint8Array(reader.result);
            };
            reader.readAsArrayBuffer(this.files[0]);
        });
}