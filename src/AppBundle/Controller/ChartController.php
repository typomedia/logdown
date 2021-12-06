<?php

namespace AppBundle\Controller;

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

    public function __construct(LogRepository $repo)
    {
        $this->repo = $repo;
    }

    /**
     * @Route("/chart")
     * @throws \Exception
     */
    public function chartAction(Request $request): Response
    {
        $logs = $this->repo->get('Chart.sql', $request);

        foreach ($logs as $log) {
            $date = date_create($log['Date']);
            $this->data['labels'][] = date_format($date,'d');
            $this->data['number'][] = $log['Number'];
            $this->data['median'][] = round($log['Average']);
        }

        return $this->render('@App/chart/index.html.twig', [
            'labels' => json_encode($this->data['labels']),
            'number' => json_encode($this->data['number']),
            'median' => json_encode($this->data['median'])
        ]);

    }
}
