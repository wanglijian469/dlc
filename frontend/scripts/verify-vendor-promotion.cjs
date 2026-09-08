/* Browser acceptance uses mocked endpoints and labeled fixtures; never writes production data.
   Usage: NODE_PATH=<directory containing playwright> node scripts/verify-vendor-promotion.cjs
   Optional: PROMOTION_BASE_URL, PROMOTION_BROWSER (executable path). */
const { chromium } = require("playwright");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const output = path.resolve(__dirname, "../../outputs/vendor-promotion");
fs.mkdirSync(output, {recursive:true});
const vendor={id:71,slug:"qavendor",name:"验收样例 · 农机液压配件厂",shortName:"验收样例厂商",province:"河北省",city:"邢台市",mainProducts:"拖拉机油缸、液压泵、农机传动配件",serviceAdvantages:"支持来图来样，提供小批量配套",description:"本页为界面自动化验收样例，不代表真实企业。",phone:"0319-0000000",phonePublic:true,wechat:"qa-vendor",wechatPublic:true,isVisible:true,publicationStatus:"published",coverImage:"/api/media/710",logo:"/api/media/711",equipment:"数控加工、装配与检测",media:[{id:1,kind:"equipment",caption:"竖版资料 · 验收样例",url:"/api/media/712"},{id:2,kind:"factory",caption:"横版资料 · 验收样例",url:"/api/media/713"}]};
const supplier={id:81,vendorId:71,vendor,productId:91,vendorProductName:"拖拉机双作用油缸",vendorModel:"QA-160",compatibleModels:"拖拉机",description:"界面验收样例：产品图片、型号和供货信息均由本厂维护。",image:"/api/media/712",gallery:["/api/media/713"],specs:[{name:"缸径",value:"80 mm"}],status:"approved",priceUnit:"件",minOrderQuantity:1,availableQuantity:100,leadTime:"7天内",freightNote:"按实际运费结算",supplyAbility:"按订单生产",updatedAt:"2026-09-07T08:00:00Z",categoryId:1};
const product={id:91,name:"公共产品分类",slug:"qa-product",categoryId:1,supplier};
const summary={days:30,vendor:{id:71,name:vendor.name,pv:0,uv:0,contacts:0,contactEvents:[],trend:[]},products:{pv:0,uv:0,trend:[],items:[]},showroom:{pv:0,uv:0,productPV:0,productUV:0,contactUV:0,conversionRate:0,contactEvents:[],sources:[],items:[]}};
const layout={siteMeta:{siteName:"大陆农机配件",brandMark:"农",mobileBrandName:"大陆农机配件",mobileBrandMark:"农",copyrightOwner:"界面验收样例",copyrightYear:"2026"},theme:{primaryColor:"#1656bd",accentColor:"#0d8b6f"},topMenus:[],sidebarMenus:[],auxiliaryMenus:[],mobileMenus:[],mobileBottomMenus:[]};
async function main(){
 const browser=await chromium.launch({headless:true,...(process.env.PROMOTION_BROWSER?{executablePath:process.env.PROMOTION_BROWSER}:{channel:"msedge"})});
 const report=[];
 try {
  for(const width of [360,390,430,1440]){
   const context=await browser.newContext({viewport:{width,height:900},deviceScaleFactor:1});
   let drafts=[],saves=0,commits=0;
   await context.addInitScript(()=>{localStorage.setItem("cms_authenticated","true");localStorage.setItem("cms_role","vendor");localStorage.setItem("cms_username","qa-vendor");});
   await context.route("**/api/**",async route=>{
    const url=new URL(route.request().url()),p=url.pathname,method=route.request().method();
    if(!p.startsWith("/api/")) return route.continue();
    if(p.startsWith("/api/media/")){
     const portrait=p.endsWith("712"),w=portrait?600:1000,h=portrait?900:550;
     return route.fulfill({contentType:"image/svg+xml",body:`<svg xmlns="http://www.w3.org/2000/svg" width="${w}" height="${h}"><rect x="8" y="8" width="${w-16}" height="${h-16}" rx="12" fill="#e4eef9" stroke="#1656bd" stroke-width="12"/><text x="30" y="55" font-size="26" fill="#1656bd">TOP · QA FULL IMAGE</text><rect x="60" y="120" width="${w-120}" height="${h-240}" fill="#b5cee9"/><text x="30" y="${h-30}" font-size="26" fill="#1656bd">BOTTOM · QA ONLY</text></svg>`});
    }
    let data;
    if(p==="/api/layout-config")data=layout;
    else if(p==="/api/site-meta")data=layout.siteMeta;
    else if(p==="/api/vendor-categories")data=[];
    else if(p==="/api/auth/session")data={username:"qa-vendor",role:"vendor",vendorId:71};
    else if(p==="/api/filter-options")data={provinces:["河北省"],categories:[{id:1,name:"液压系统配件",isEnabled:true}],serviceTags:[]};
    else if(p==="/api/admin/vendor-workspace")data={vendor,counts:{pending:1,approved:1,rejected:0,expired:0,missingImage:0},draftCount:drafts.length,unreadCount:1,profileStatus:"approved"};
    else if(p==="/api/admin/vendor-analytics")data=summary;
    else if(p==="/api/admin/vendor-products")data=[{...supplier,recordType:"supplier",supplierId:81,priceVersion:1}];
    else if(p.endsWith("/duplicate-check"))data={exact:false,similar:[]};
    else if(p==="/api/admin/vendor-work-drafts")data=drafts;
    else if(p.startsWith("/api/admin/vendor-work-drafts/")&&p.endsWith("/commit")){commits++;data={...drafts[0],committedAt:new Date().toISOString()};drafts=[];}
    else if(p.startsWith("/api/admin/vendor-work-drafts/")&&method==="PUT"){saves++;const input=route.request().postDataJSON();data={id:1,...input,clientKey:p.split("/").at(-1),version:input.version+1,updatedAt:new Date().toISOString()};drafts=[data];}
    else if(p==="/api/vendors/slug/qavendor"||p==="/api/vendors/71")data=vendor;
    else if(p==="/api/vendors/71/posts")data=[];
    else if(p==="/api/showrooms/qavendor/products/81")data=supplier;
    else if(p==="/api/showrooms/qavendor/products")data={items:[product],page:1,pageSize:12,total:1};
    else if(p==="/api/analytics/events")data={recorded:false};
    else if(p.startsWith("/api/seo"))data={title:"界面验收样例",description:"",canonical:""};
    else return route.fulfill({status:404,json:{code:404,message:"QA endpoint not mocked"}});
    return route.fulfill({json:{code:0,data}});
   });
   const page=await context.newPage();const errors=[];page.setDefaultTimeout(10000);page.on("console",m=>{if(m.type()==="error")console.error(m.text())});page.on("pageerror",e=>errors.push(e.message));
   const base=process.env.PROMOTION_BASE_URL||"http://127.0.0.1:5173";
   const shot=async name=>{await page.screenshot({path:path.join(output,`${name}-${width}.png`),fullPage:false,animations:"disabled"});const overflow=await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth+1);assert.equal(overflow,false,name+" horizontal overflow at "+width);};
   await page.goto(base+"/admin/vendor-workspace");await page.getByRole("heading",{name:"验收样例厂商",exact:true}).waitFor().catch(async e=>{await page.screenshot({path:path.join(output,"failure.png")});console.log(await page.locator("body").innerText());console.log(errors);throw e});await shot("workspace");
   await page.getByRole("link",{name:"发布产品",exact:true}).first().click();await page.getByRole("dialog",{name:"添加本厂产品"}).waitFor();await shot("entry-step1");
   await page.getByLabel("本厂产品名称",{exact:true}).fill("恢复测试油缸");await page.getByLabel("本厂型号",{exact:true}).fill("QA-161");await page.getByLabel("产品大类").selectOption("1");
   await page.getByRole("button",{name:"保存草稿",exact:true}).click();await page.getByText("已保存",{exact:true}).waitFor();
   assert.ok(saves>0);await page.getByRole("button",{name:"取消",exact:true}).click();await page.getByRole("button",{name:"继续编辑",exact:true}).click();assert.equal(await page.getByLabel("本厂型号",{exact:true}).inputValue(),"QA-161");
   await page.getByRole("button",{name:"下一步",exact:true}).click();assert.equal(await page.getByLabel("库存 / 可供应量",{exact:true}).inputValue(),"100");await shot("entry-step2");
   await page.getByRole("button",{name:"下一步",exact:true}).click();await shot("entry-step3");await page.getByLabel("我已确认产品资料及供货信息真实准确").check();await page.getByRole("button",{name:"确认并提交审核",exact:true}).click();await page.getByText("产品资料已提交平台审核",{exact:true}).waitFor();assert.equal(commits,1);
   await page.goto(base+"/v/qavendor");await page.getByRole("heading",{name:vendor.name,exact:true,level:1}).waitFor();await page.getByText(supplier.vendorProductName,{exact:true}).waitFor();await shot("showroom");
   await page.getByText(supplier.vendorProductName,{exact:true}).click();await page.getByRole("button",{name:"放大本厂产品主图"}).waitFor();assert.match(page.url(),/\/v\/qavendor\/products\/81/);await shot("own-product");
   await page.getByRole("button",{name:"放大本厂产品主图"}).click();await page.getByRole("dialog",{name:"图片预览"}).waitFor();await shot("lightbox");await page.keyboard.press("Escape");
   await page.getByRole("button",{name:"分享推广"}).click();await page.locator(".promotion-qr").waitFor();await shot("promotion");await page.getByRole("button",{name:"生成海报",exact:true}).click();await page.locator(".promotion-poster").waitFor();assert.match(await page.getByRole("link",{name:"保存推广海报"}).getAttribute("href"),/^data:image\/png;base64,/);await page.locator(".promotion-poster").screenshot({path:path.join(output,`poster-${width}.png`)});
   assert.deepEqual(errors,[]);report.push({width,saves,commits,errors});await context.close();
  }
 } finally {await browser.close();}
 fs.writeFileSync(path.join(output,"acceptance.json"),JSON.stringify(report,null,2));console.log(JSON.stringify(report));
}
main().catch(e=>{console.error(e);process.exitCode=1;});
