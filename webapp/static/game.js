var renderer = null,
	stage = null,
	selfSprite = null,
	SPEED = 0.2;
	websockPort = "",
	spriteMap = {},
	textMap = {}, //maps playerName -> text object
	selfNodeId = -1,
	selfSpriteInfo = {},
	colorMap = {},
	//ensures self dead update is sent only once
	sentSelfDeadUpdate = false;
	lastAjaxTime = -1;
	selfPlayerName = "",
	htmlBodyElement=null,
	prevArrowPress = "",
	currZone = -1,
	pelletSprites = [],
	spriteZoneMap = {};

//sprite dict related constants
const PLAYERNAME_KEY = "PlayerName"
const DIRECTION_KEY = "Direction"
const ZONE_KEY = "Zone"
const ZONE_CHANGED_KEY = "ZoneChanged"
const XVEL_KEY = "Xvel"
const YVEL_KEY = "Yvel"
const XSIZE_KEY = "Xsize"
const YSIZE_KEY = "Ysize"
const XPOS_KEY = "Xpos"
const YPOS_KEY = "Ypos"
const COLOR_KEY = "Color"
const ISALIVE_KEY = "IsAlive"
const PLANET_IMGS = ["mercury.png", "venus.png", "earth.png", "mars.png", "jupiter.png", "saturn.png", "uranus.png", "neptune.png", "pluto.png"]

const DEBUG_PANEL_ID = "debug-panel"
const LABEL_TEXT_FONT = "Ariel"
const LABEL_TEXT_SIZE = 20
const LABEL_TEXT_COLOR = "white"
const SCREEN_WIDTH = 700
const SCREEN_HEIGHT = 500
const FOOD_PELLET_REFRESH_RATE_MS = 4000
const PERIODIC_UPDATE_PERIOD = 3000 //secs

//minimum interval between 2 Ajax updates
const AJAX_UPDATE_INTERVAL_MS = 400;

function main() {
	console.log("Hello hello");
	htmlBodyElement = document.getElementById("body");
	var type = "WebGL"
    if(!PIXI.utils.isWebGLSupported()){
      type = "canvas"
    }

    PIXI.utils.sayHello(type)
	ajaxGetWebsockPort();
    pixiSetup();
}

function pixiSetup() {
	//Create the renderer
	renderer = PIXI.autoDetectRenderer(SCREEN_WIDTH, SCREEN_HEIGHT, 
		{antialias: false, transparent: false, resolution: 1});

	renderer.view.style.position = "absolute";
	renderer.view.style.display = "block";
	// renderer.autoResize = true;
	// renderer.resize(window.innerWidth, window.innerHeight);

	//Add the canvas to the HTML document
	document.body.appendChild(renderer.view);

	//Create a container object called the `stage`
	stage = new PIXI.Container();

	//Tell the `renderer` to `render` the `stage`
	renderer.render(stage);
	console.log(" ^^ Calling ajaxGetSelfBlobInfo.");
	ajaxGetSelfBlobInfo();

	PIXI.loader.add("static/img/circ.png")
	.add("static/img/mercury.png")
	.add("static/img/venus.png")
	.add("static/img/earth.png")
	.add("static/img/mars.png")
	.add("static/img/jupiter.png")
	.add("static/img/saturn.png")
	.add("static/img/uranus.png")
	.add("static/img/neptune.png")
	.add("static/img/pluto.png")
	.load(loadComplete);
}

function loadComplete() {

	//Render the stage   
	renderer.render(stage);

	window.setInterval(addFoodPellet, FOOD_PELLET_REFRESH_RATE_MS);
	gameLoop();

  	window.addEventListener("keydown", keyCapture, true);

  	//initialize websocket
	if(websockPort.length < 2) {
		console.log(" !! Error: websock port is not valid.");
		return;
	}
  	var ws = new WebSocket("ws://localhost:" + websockPort + "/subscribe", "protocolOne");
    ws.onopen  = function (event) {
    	console.log("ws onopen!");
	  	// ws.send("Here's some text that the server is urgently awaiting!"); 
	}; 

	ws.onmessage = onServerUpdate

	ws.onclose = function(event) {
		console.log("Websocket connection closed.");
	};

	ws.onerror = function (error) {
      console.log('Websocket error:' + JSON.stringify(error));
    };

}


function onServerUpdate(event) {
	var blobList = JSON.parse(event.data);

	for(var i=0; i<blobList.length; i++) {
		var updateDict = blobList[i];
		console.log(" -- Got update: " + updateDict[PLAYERNAME_KEY] + ", " +
			 updateDict[ZONE_KEY] + "(" + updateDict[XPOS_KEY] + ", " + updateDict[YPOS_KEY] + ")");
		var playerName = updateDict[PLAYERNAME_KEY];
		colorMap[playerName] = updateDict[COLOR_KEY]
		spriteZoneMap[playerName] = updateDict[ZONE_KEY]
		if(playerName in spriteMap) {
			spriteFromDict(spriteMap, updateDict);
			if(!updateDict[ISALIVE_KEY]) {
				//destroy sprite
				var deadSprite = spriteMap[playerName];
				console.log("Destroying sprite: " + deadSprite.cursor)
				deadSprite.width = 1;
				deadSprite.height = 1;
				var deadSpriteLabel = textMap[deadSprite.cursor];
				deadSpriteLabel.text = ":(";
			}
		} else {
			//create sprite, add to stage
			var newSprite = createSprite(updateDict);
			spriteMap[playerName] = newSprite;
			stage.addChild(newSprite);

			//create label, add to stage
			var label = new PIXI.Text(getDisplayName(newSprite.cursor),
				{fontFamily: LABEL_TEXT_FONT, fontSize: LABEL_TEXT_SIZE, fill: LABEL_TEXT_COLOR}
			);

			label.position.set(newSprite.x, newSprite.y);
			label.anchor.x = 0.5;
			label.anchor.y = 0.5;
			textMap[newSprite.cursor] = label;
			stage.addChild(label);
			//render new stage
			renderer.render(stage);
		}
	}
}

function gameLoop() {
  //Loop this function at 60 frames per second
  
  requestAnimationFrame(gameLoop);
  updateSpritePositions();
  
  checkCollisions();
  checkPelletConsumed();
  checkZoneChanged();
  //Render the stage to see the animation
  renderer.render(stage);
}

function keyCapture(event) {
	if(event.key == prevArrowPress) {
		return;
	}
	var selfSpr = spriteMap[selfPlayerName];
	switch(event.key) {
		case "ArrowUp":
			selfSpr.vx = 0;
			selfSpr.vy = -1*SPEED;
			event.preventDefault();
			break;
		case "ArrowDown":
			selfSpr.vx = 0;
			selfSpr.vy = SPEED;
			event.preventDefault();
			break;
		case "ArrowLeft":
			selfSpr.vy = 0;
			selfSpr.vx = -1*SPEED;
			event.preventDefault();
			break;
		case "ArrowRight":
			selfSpr.vy = 0;
			selfSpr.vx = SPEED;
			event.preventDefault();
			break;
		default:
			return;
	}
	prevArrowPress = event.key;
	var newStateJson = JSON.stringify(spriteToMap(selfSpr))
	ajaxNodeStateUpdate(newStateJson)
}

function ajaxGetWebsockPort() {
	var req = new XMLHttpRequest();
    req.onreadystatechange = function() {
     	console.log("Got websocket port!: " + req.responseText);	   
		websockPort = req.responseText;
    };

    req.open("GET", "/websocketport", true);
	lastAjaxTime = (new Date()).getTime();
    req.send(); 
}

function ajaxGetSelfBlobInfo() {
	var req = new XMLHttpRequest();
    req.onreadystatechange = function() {
     	if(selfPlayerName.length > 0) {
			console.log(" -- Returning; already been called.");
			return;
		}  
		selfNodeInfoJson = req.responseText;
		if(selfNodeInfoJson.length <= 0) {
			console.log(" -- incorrect response for selfNodeId, returning.");
			return;
		}

		var jsonDict = JSON.parse(selfNodeInfoJson);
		selfPlayerName = jsonDict[PLAYERNAME_KEY];
		selfSpriteColor = jsonDict[COLOR_KEY];
		currZone = jsonDict[ZONE_KEY];

    	document.getElementById("zone-img").src = "static/img/q-" + currZone + ".png";
		console.log("curr zone: " + currZone)
		colorMap[selfPlayerName] = selfSpriteColor
		spriteZoneMap[selfPlayerName] = jsonDict[ZONE_KEY]
		//create new sprite
		var circGraphic = new PIXI.Graphics();
		circGraphic.beginFill(jsonDict[COLOR_KEY]);
		circGraphic.drawCircle(40, 60, 1000);
		circGraphic.endFill();
		var circleTexture = circGraphic.generateTexture(); 
		var newSprite = new PIXI.Sprite(circleTexture);
		newSprite.anchor.x = 0.5;
		newSprite.anchor.y = 0.5;
		// Add the sprite to the stage
		stage.addChild(newSprite);
		spriteMap[jsonDict[PLAYERNAME_KEY]] = newSprite;
		spriteFromDict(spriteMap, jsonDict);

		var message = new PIXI.Text(getDisplayName(selfPlayerName),
			{fontFamily: LABEL_TEXT_FONT, fontSize: LABEL_TEXT_SIZE, fill: LABEL_TEXT_COLOR}
		);

		message.anchor.x = 0.5;
		message.anchor.y = 0.5;
		message.position.set(newSprite.x, newSprite.y);
		stage.addChild(message);
		textMap[newSprite.cursor] = message
		//Render the stage 
		renderer.render(stage);
		//broadcast update to all nodes
		var newStateJson = JSON.stringify(
			spriteToMap(spriteMap[selfPlayerName]))
		ajaxNodeStateUpdate(newStateJson)
    };

    req.open("GET", "/selfnodeinfo", true);
    req.send(); 
}

window.onload = main
//TODO resize renderer on window resize

function getDisplayName(realName) {
	return realName	
	//var arr = realName.split("|")
	// return arr[0]
}


function periodicStateUpdate() {
	var selfSpr = spriteMap[selfPlayerName]
	var newStateJson = JSON.stringify(spriteToMap(selfSpr))
	ajaxNodeStateUpdate(newStateJson)
}