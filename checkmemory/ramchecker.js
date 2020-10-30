const os = require('os');
const disk = require('diskusage');

let free = null;
const CheckMemory = () =>{
    const total_memory = os.totalmem(); 
    const free_memory = os.freemem();
    const used_memory = total_memory - free_memory;
    const usedmemper = used_memory * 100 / total_memory; 
    return usedmemper.toFixed(2)+" %";
}

const CheckDisk = async () =>{
    await disk.check('/', function(err, info) {
        const Total_space = info.total;
        const free_space = info.free ;
        const used_space = Total_space - free_space;
        const usedDiskper = used_space * 100 / Total_space;
        free = usedDiskper.toFixed(2);
        return 
    });
}


const Warning = (mem) =>{
    // console.log(free+ "used disk per");
    let i = parseFloat(mem);
    let j = parseFloat(free);
    // console.log(free)
    // console.log(i)
    if (i >=  80) {
        console.log("You are currently using 80% of your available memory");
    }
    if(j >= 80){
        console.log("You are currently using 80% of your available disk");
    }
} 

async function main(){
    const Usedram = await CheckMemory();
    await CheckDisk();
    await setTimeout(async () => {await Warning(Usedram)},2000)

}

main();
