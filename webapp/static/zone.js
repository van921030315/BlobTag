

function checkZoneChanged() {
    var selfSprite = spriteMap[selfPlayerName];
    if(!isSpriteAlive(selfSprite)) {
        return;
    }

    if(selfSprite.x > SCREEN_WIDTH + 10) {
        if(currZone == 1) {
            selfSprite.x = 0;
            //zone update to 0
            zoneUpdate(0);
            
        } else if(currZone == 2) {
            selfSprite.x = 0;
            //zone updateto 3
            zoneUpdate(3);
        }
        return
    } else if(selfSprite.x < -5) {

        if(currZone == 0) {
            selfSprite.x = SCREEN_WIDTH;
            //z update to 1
            zoneUpdate(1);
        } else if(currZone == 3) {
            //z update to 2
            selfSprite.x = SCREEN_HEIGHT;
            zoneUpdate(2);
        }
        return
    } else if(selfSprite.y > SCREEN_HEIGHT) {
        if(currZone == 1) {
            selfSprite.y = 0;
            //upd to 2
            zoneUpdate(2);
        } else if(currZone == 0) {
            //upd to 3
            selfSprite.y = 0;
            zoneUpdate(3);
        }
        return
    } else if(selfSprite.y < -5) {
        if(currZone == 2) {
            //upd to 1
            selfSprite.y = SCREEN_HEIGHT;
            zoneUpdate(1);
        } else if(currZone == 3) {
            //upd to 0
            selfSprite.y = SCREEN_HEIGHT;
            zoneUpdate(0);
        }
        return
    }
}


function zoneUpdate(newZone) {
    currZone = newZone;
    var selfSprite = spriteMap[selfPlayerName]
    spriteZoneMap[selfPlayerName] = newZone
    var updateMap = spriteToMap(selfSprite);
    updateMap[ZONE_CHANGED_KEY] = true
    var newStateJson = JSON.stringify(updateMap);
    ajaxNodeStateUpdate(newStateJson);
    resetAllSpriteStates();
    document.getElementById("zone-img").src = "static/img/q-" + newZone + ".png";
}

//removes all sprite except selfsprite
function resetAllSpriteStates() {
    //sprites
    var selfSprite = spriteMap[selfPlayerName];
    for(var plName in spriteMap) {
        if(plName == selfPlayerName) {
            continue;
        }
        var sprite = spriteMap[plName];
        sprite.destroy();
    }
    spriteMap = {};
    spriteMap[selfPlayerName] = selfSprite;
    //zoneMap
    spriteZoneMap = {};
    spriteZoneMap[selfPlayerName] = currZone;
    
    //textMap
    var selfLabel = textMap[selfPlayerName];
    for(var plName in textMap) {
        if(plName == selfPlayerName) {
            continue
        }
        var currLabel = textMap[plName];
        currLabel.destroy();
    }
    textMap = {};
    textMap[selfPlayerName] = selfLabel;

    //colorMap
    var selfColor = colorMap[selfPlayerName];
    colorMap = {};
    colorMap[selfPlayerName] = selfColor;
    //pellets
    for(var i=0; i<pelletSprites.length; i++) {
        var pelletSprite = pelletSprites[i];
        pelletSprite.destroy()
    }
    pelletSprites = []
}
