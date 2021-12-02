select
    Date,
    Request,
    Param,
    count (*) as Number,
    avg (Duration) as Average
from Logs
--where Param = 'type=SoapAiDAlerts'
--where Param = 'type=SoapXML'
where Request = :request
and date like :date
group by Date, Request
order by Date, Number desc;