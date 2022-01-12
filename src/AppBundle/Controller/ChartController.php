<?php

namespace AppBundle\Controller;

use AppBundle\Helper\DateHelper;
use AppBundle\Repository\LogRepository;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class ChartController extends AbstractController
{
    /**
     * @var array $data
     */
    protected $data;

    /**
     * @var LogRepository $repo
     */
    private $repo;

    /**
     * @param LogRepository $repo
     */
    public function __construct(LogRepository $repo)
    {
        $this->repo = $repo;
    }

    /**
     * @Route("/chart")
     */
    public function chart(Request $request): Response
    {
        $logs = $this->repo->get('Chart.sql');
        $view = DateHelper::analyzeDate($request->query->get('date'));

        // prepare data for chartist.js
        foreach ($logs as $log) {
            $date = date_create($log['datetime']);
            if ($view) {
                if ($view === 'day') {
                    $date = date_format($date, 'H');
                } elseif ($view === 'hour') {
                    $date = date_format($date, 'H:i');
                } else { // month
                    $date = date_format($date, 'd');
                }
            }
            $this->data['labels'][] = $date;
            $this->data['number'][] = $log['number'];
            $this->data['median'][] = round($log['average']);
        }

        $info = $this->repo->get('Chart.sql', 1);

        return $this->render('@App/sites/chart.html.twig', [
            'info'   => $info,
            'view'   => $view,
            'labels' => json_encode($this->data['labels']),
            'number' => json_encode($this->data['number']),
            'median' => json_encode($this->data['median'])
        ]);
    }
}
