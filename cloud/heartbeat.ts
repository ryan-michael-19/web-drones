import { SNSClient, PublishCommand, KMSOptInRequired } from "@aws-sdk/client-sns";
import { Handler } from 'aws-lambda';
import process from 'process';

export const handler: Handler = async (event, context) => {
    await runHeartbeat(event["snsArn"])
}

async function runHeartbeat(snsArnForFailure: string) {
    const snsClient = new SNSClient({});
    const heartbeatStart = new Date();
    try {
        const res = await fetch("https://webdrones.net/");
        if (res.status === 200) {
            console.log(`Heartbeat successful at ${heartbeatStart}`);
            return "success";
        } else {
            console.log(`Request complete but heartbeat failed" at ${heartbeatStart}`);
            const snsRes = await snsClient.send(
                new PublishCommand({
                    Message: `Soft heartbeat fail at ${heartbeatStart}`,
                    TopicArn: snsArnForFailure,
                })
            );
            return "soft fail";
        }
    } catch {
        console.log(`Request incomplete so heartbeat failed" at ${heartbeatStart}`);
        const snsRes = await snsClient.send(
            new PublishCommand({
                Message: `Hard heartbeat fail at ${heartbeatStart}`,
                TopicArn: snsArnForFailure,
            })
        );
        return "hard fail"
    }
}

if (require.main === module) {
    (async ()=> {
        await runHeartbeat(process.argv[2])
    })();
}