import { OpenAPI, DefaultService } from "./api";

OpenAPI.BASE = "/api"

// #############################################################
// npm run dev, navigate to client url and check browser console
// #############################################################

await DefaultService.postHello()

const answer = await DefaultService.getAnswer()
console.log(answer)

const item = await DefaultService.postItemsByMenuId({ menuId: 1, requestBody: { Name: "updatedItem", Price: 22.2 } })
console.log(item)

await DefaultService.putItems({ requestBody: { Name: "newItem", Price: 33.3 } })
await DefaultService.deleteMenusById({ id: 2, quiet: false })
