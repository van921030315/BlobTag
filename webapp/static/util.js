

function spriteToMap(sprite, color) {
    returnMap = {}
    returnMap[ISALIVE_KEY] = !sentSelfDeadUpdate;
    returnMap[XVEL_KEY] = sprite.vx;
    returnMap[YVEL_KEY] =  sprite.vy;
    returnMap[XPOS_KEY] = sprite.x;
    returnMap[YPOS_KEY] = sprite.y;
    returnMap[XSIZE_KEY] = sprite.width;
    returnMap[YSIZE_KEY] = sprite.height;
    returnMap[PLAYERNAME_KEY] = sprite.cursor;
    returnMap[COLOR_KEY] = colorMap[sprite.cursor];
    returnMap[ZONE_KEY] = spriteZoneMap[sprite.cursor];
    returnMap[ZONE_CHANGED_KEY] = false
    return returnMap
}

function spriteFromDict(spriteMap, jsonDict) {
    var playerName = jsonDict[PLAYERNAME_KEY]
    var sprite = spriteMap[playerName]
    if(sprite == undefined) {
        console.log(" Undefined sprite; creating new...");
        return
    } else if(!isSpriteAlive(sprite)) {
        return;
    }
    sprite.cursor = playerName;
    sprite.vx = jsonDict[XVEL_KEY];
    sprite.vy = jsonDict[YVEL_KEY];
    sprite.x = jsonDict[XPOS_KEY];
    sprite.y = jsonDict[YPOS_KEY];
    sprite.width = jsonDict[XSIZE_KEY];
    sprite.height = jsonDict[YSIZE_KEY];
}

function addFoodPellet() {
    var planetIndex = Math.floor(Math.random() * 1000 % 9);
    var pelletSprite = new PIXI.Sprite(
        PIXI.loader.resources["static/img/" + PLANET_IMGS[planetIndex]].texture
    );
    pelletSprite.width = 10;
    pelletSprite.height = 10;
    pelletSprite.x = Math.floor(Math.random() * (SCREEN_WIDTH * 0.9) + 50);
    pelletSprite.y = Math.floor(Math.random() * (SCREEN_HEIGHT) + 50);
    pelletSprite.anchor.x = 0.5;
    pelletSprite.anchor.y = 0.5;
    stage.addChild(pelletSprite);
    pelletSprites.push(pelletSprite);
}

function ajaxNodeStateUpdate(jsonUpdateStr) {
	var req = new XMLHttpRequest();
    req.onreadystatechange = function() {
     	// console.log("Got response for state update!: " + req.responseText);	   
    };

    req.open("GET", "/nodestateupdate/newstate?state=" + jsonUpdateStr, true);
    console.log(" -- Sending Ajax node state update.")
    var currEpochTime = (new Date()).getTime();
    if(currEpochTime - lastAjaxTime < AJAX_UPDATE_INTERVAL_MS) {
        window.setTimeout(function () { req.send(); }, AJAX_UPDATE_INTERVAL_MS)
    } else {
        req.send();
    } 
}

function createSprite(jsonDict) {
    var playerName = jsonDict[PLAYERNAME_KEY];
    //create new sprite
    var circGraphic = new PIXI.Graphics();
    circGraphic.beginFill(jsonDict[COLOR_KEY]);
    circGraphic.drawCircle(40, 60, 1000);
    circGraphic.endFill();
    var circleTexture = circGraphic.generateTexture(); 
    var newSprite = new PIXI.Sprite(circleTexture);

    newSprite.cursor = playerName;
    newSprite.vx = jsonDict[XVEL_KEY];
    newSprite.vy = jsonDict[YVEL_KEY];
    newSprite.x = jsonDict[XPOS_KEY];
    newSprite.y = jsonDict[YPOS_KEY];
    newSprite.width = jsonDict[XSIZE_KEY];
    newSprite.height = jsonDict[YSIZE_KEY];
	newSprite.anchor.x = 0.5;
	newSprite.anchor.y = 0.5;
	newSprite.cursor = playerName;
    return newSprite;
}

function updateSpritePositions() {
    var debugPanel = document.getElementById(DEBUG_PANEL_ID);
    var debugMsg = "Zone " + currZone + "; ";
    for(var key in spriteMap) {
        var currSprite = spriteMap[key];
        if(currSprite == null) {
            console.log("!! currSprite null");
            continue;
        } 
        currSprite.x += currSprite.vx;
        currSprite.y += currSprite.vy;
        debugMsg += currSprite.cursor + "(" + currSprite.x.toFixed() + "," + currSprite.y.toFixed() + ")   ";
        //set text position
        var currText = textMap[currSprite.cursor];
        if(currText != null) {
            currText.position.set(currSprite.x, currSprite.y);
        }
    }
    debugPanel.innerHTML = debugMsg;
}

function checkPelletConsumed() {
    var deadPelletIndices = {};

    //check for pellet-sprite collisions
    for(var playerName in spriteMap) {
        var currSprite = spriteMap[playerName];
        if(!isSpriteAlive(currSprite)) {
            continue;
        }
        for(var i=0; i<pelletSprites.length; i++) {
            var pelletSprite = pelletSprites[i];
            if(checkContact(currSprite, pelletSprite)) {
                if(playerName == selfPlayerName) {
                    //increase size; send update
                    increaseSpriteSizeBy(currSprite, 700);
                    var newStateJson = JSON.stringify(spriteToMap(currSprite))
                    ajaxNodeStateUpdate(newStateJson)
                }
                //kill pellet
                deadPelletIndices[i] = true;
            }
        }

        for(var deadPelletIdx in deadPelletIndices) {
            var deadPelletSprite = pelletSprites[deadPelletIdx];
            if(deadPelletSprite == undefined || deadPelletSprite == null) {
                continue;
            }
            //kill
            deadPelletSprite.destroy();
            //remove from list
            pelletSprites.splice(deadPelletIdx, 1);
        }
    }
}

//check if self sprite collided with any other sprite
function checkCollisions() {
    var selfSprite = spriteMap[selfPlayerName];
    if(!isSpriteAlive(selfSprite)) {
        return;
    }
    for(var playerName in spriteMap) {
        if(playerName == selfPlayerName) {
            continue;
        }
        //check for collosion
        var otherSprite = spriteMap[playerName];
        if(checkContact(selfSprite, otherSprite) && 
            isSpriteAlive(selfSprite) && isSpriteAlive(otherSprite)) {
            //equal size case: pass through each other
            if(selfSprite.width == otherSprite.width) {
                continue;
            }
            if(selfSprite.width + selfSprite.height > 
                        otherSprite.width + otherSprite.height) {
                //kill other sprite
                increaseSpriteSizeBy(selfSprite, getSpriteSize(otherSprite) * 0.75);
                otherSprite.width = 1;
                otherSprite.height = 1;
                var otherLabel = textMap[otherSprite.cursor];
                otherLabel.text = ":("
                console.log(otherSprite.cursor + " died.");
                updateDict = spriteToMap(otherSprite);
                updateDict[ISALIVE_KEY] = false;
                ajaxNodeStateUpdate(JSON.stringify(updateDict))
                
            } else {
                //kill self sprite
                if(sentSelfDeadUpdate) {
                    break;
                }
                updateDict = spriteToMap(selfSprite);
                updateDict[ISALIVE_KEY] = false;
                ajaxNodeStateUpdate(JSON.stringify(updateDict))
                sentSelfDeadUpdate = true;
                console.log(selfSprite.cursor + " died.");
            }
        }
    }
}


/**
 * Change sprite width, height in order to increase it area by incr
 */
function increaseSpriteSizeBy(spr, incr) {
    //calculate new area

    var newSize = (Math.PI * spr.width * spr.width) + incr;
    var newRadius = Math.sqrt(newSize / Math.PI);
    spr.width = newRadius;
    spr.height = newRadius;
}

function getSpriteSize(spr) {
    return Math.PI * spr.width * spr.width;
}

/**
 * Check if circular sprites spr1 and spr2 are in contact
 */
function checkContact(spr1, spr2) {
    var centerDist = Math.sqrt(Math.pow(spr1.x - spr2.x, 2) + 
                    Math.pow(spr1.y - spr2.y, 2));
    return centerDist < (spr1.width + spr2.width) * 0.5;    
}

function isSpriteAlive(spr) {
    if(spr == undefined || spr == null) {
        console.error("!!! spr was null in isSpriteAlive:")
        return false
    }
    return spr.width > 2 && spr.height > 2;
}